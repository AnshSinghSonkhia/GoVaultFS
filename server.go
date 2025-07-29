// File server implementation for GoVaultFS
// This file defines the distributed file server node, its network protocol, and file operations.
// Each node can store, retrieve, and replicate files across a peer-to-peer network.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/AnshSinghSonkhia/GoVaultFS/p2p"
)

// FileServerOpts holds configuration for a file server node
type FileServerOpts struct {
	ID                string            // Unique node identifier
	EncKey            []byte            // AES encryption key
	StorageRoot       string            // Local storage directory
	PathTransformFunc PathTransformFunc // Hash-to-path converter
	Transport         p2p.Transport     // Network transport layer
	BootstrapNodes    []string          // List of bootstrap peer addresses
}

// FileServer represents a node in the distributed file system
type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex          // Protects concurrent access to peers map
	peers    map[string]p2p.Peer // Connected peer nodes

	store  *Store        // Local file storage
	quitch chan struct{} // Channel to signal server shutdown
}

// NewFileServer creates a new file server node with the given options
func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	// Generate a unique ID if not provided
	if len(opts.ID) == 0 {
		opts.ID = generateID()
	}

	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

// broadcast sends a message to all connected peers
func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send([]byte{p2p.IncomingMessage}) // Signal incoming message
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

// Message is a generic wrapper for network messages
type Message struct {
	Payload any // Can be MessageStoreFile or MessageGetFile
}

// MessageStoreFile requests a peer to store a file
type MessageStoreFile struct {
	ID   string // Node ID
	Key  string // File hash
	Size int64  // File size
}

// MessageGetFile requests a peer to send a file
type MessageGetFile struct {
	ID  string // Node ID
	Key string // File hash
}

// Get retrieves a file by key.
// If the file is not found locally, it requests it from peers and stores the result locally.
func (s *FileServer) Get(key string) (io.Reader, error) {
	// Check if file exists locally
	if s.store.Has(s.ID, key) {
		fmt.Printf("[%s] serving file (%s) from local disk\n", s.Transport.Addr(), key)
		_, r, err := s.store.Read(s.ID, key)
		return r, err
	}

	// File not found locally, request from peers
	fmt.Printf("[%s] dont have file (%s) locally, fetching from network...\n", s.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			ID:  s.ID,
			Key: hashKey(key),
		},
	}

	// Broadcast request to all peers
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	// Wait for peers to respond
	time.Sleep(time.Millisecond * 500)

	// Receive file from peers
	for _, peer := range s.peers {
		// Read file size from peer
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)

		// Decrypt and write file to local storage
		n, err := s.store.WriteDecrypt(s.EncKey, s.ID, key, io.LimitReader(peer, fileSize))
		if err != nil {
			return nil, err
		}

		fmt.Printf("[%s] received (%d) bytes over the network from (%s)\n", s.Transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}

	// Return file reader from local storage
	_, r, err := s.store.Read(s.ID, key)
	return r, err
}

// Store saves a file locally and replicates it to all peers.
// The file is encrypted before storage and transfer.
func (s *FileServer) Store(key string, r io.Reader) error {
	var (
		fileBuffer = new(bytes.Buffer)           // Buffer to hold file data for replication
		tee        = io.TeeReader(r, fileBuffer) // TeeReader writes to buffer and local storage simultaneously
	)

	// Write file to local storage
	size, err := s.store.Write(s.ID, key, tee)
	if err != nil {
		return err
	}

	// Notify peers to prepare for incoming file
	msg := Message{
		Payload: MessageStoreFile{
			ID:   s.ID,
			Key:  hashKey(key),
			Size: size + 16, // Add padding for encryption
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	// Short delay to allow peers to prepare
	time.Sleep(time.Millisecond * 5)

	// Send encrypted file to all peers
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}
	mw := io.MultiWriter(peers...)
	mw.Write([]byte{p2p.IncomingStream}) // Signal incoming stream
	n, err := copyEncrypt(s.EncKey, fileBuffer, mw)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] received and written (%d) bytes to disk\n", s.Transport.Addr(), n)

	return nil
}

// Stop signals the file server to shut down
func (s *FileServer) Stop() {
	close(s.quitch)
}

// OnPeer is called when a new peer connects
func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p // Add peer to map

	log.Printf("connected with remote %s", p.RemoteAddr())

	return nil
}

// loop is the main event loop for the file server
// It processes incoming RPCs and handles shutdown
func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to error or user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			// Decode incoming message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println("decoding error: ", err)
			}
			// Handle the message
			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println("handle message error: ", err)
			}

		case <-s.quitch:
			return
		}
	}
}

// handleMessage dispatches incoming messages to the correct handler
func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}

	return nil
}

// handleMessageGetFile serves a file to a requesting peer
func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	// Check if file exists locally
	if !s.store.Has(msg.ID, msg.Key) {
		return fmt.Errorf("[%s] need to serve file (%s) but it does not exist on disk", s.Transport.Addr(), msg.Key)
	}

	fmt.Printf("[%s] serving file (%s) over the network\n", s.Transport.Addr(), msg.Key)

	fileSize, r, err := s.store.Read(msg.ID, msg.Key)
	if err != nil {
		return err
	}

	// Close file after sending if possible
	if rc, ok := r.(io.ReadCloser); ok {
		fmt.Println("closing readCloser")
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not in map", from)
	}

	// Send stream signal and file size
	peer.Send([]byte{p2p.IncomingStream})
	binary.Write(peer, binary.LittleEndian, fileSize)
	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] written (%d) bytes over the network to %s\n", s.Transport.Addr(), n, from)

	return nil
}

// handleMessageStoreFile receives and stores a file sent by a peer
func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	// Write file to local storage
	n, err := s.store.Write(msg.ID, msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	fmt.Printf("[%s] written %d bytes to disk\n", s.Transport.Addr(), n)

	peer.CloseStream()

	return nil
}

// bootstrapNetwork connects to all bootstrap peers
func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Printf("[%s] attemping to connect with remote %s\n", s.Transport.Addr(), addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}

	return nil
}

// Start launches the file server: listens for connections, bootstraps peers, and enters event loop
func (s *FileServer) Start() error {
	fmt.Printf("[%s] starting fileserver...\n", s.Transport.Addr())

	// Start listening for incoming connections
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	// Connect to bootstrap peers
	s.bootstrapNetwork()

	// Enter main event loop
	s.loop()

	return nil
}

// init registers message types for gob encoding/decoding
func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}
