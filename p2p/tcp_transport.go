// TCP transport implementation for GoVaultFS P2P networking
// This file provides types and logic for peer management, connection handling, and message passing over TCP.
package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represents a remote node connected via TCP.
// It wraps the net.Conn and tracks whether the connection is outbound (initiated by us) or inbound (accepted from another node).
// The WaitGroup is used for synchronizing stream operations (e.g., file transfers).
type TCPPeer struct {
	net.Conn                 // Underlying TCP connection
	outbound bool            // True if connection was dialed (outbound), false if accepted (inbound)
	wg       *sync.WaitGroup // Used to block/unblock stream operations
}

// NewTCPPeer creates a new TCPPeer instance for a given connection and direction.
func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

// CloseStream signals that a stream operation (e.g., file transfer) is complete for this peer.
func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

// Send writes a byte slice to the peer's TCP connection.
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

// TCPTransportOpts holds configuration for TCPTransport.
//
//	ListenAddr    - Address to listen for incoming connections
//	HandshakeFunc - Function to run on new peer connections (e.g., authentication)
//	Decoder       - Message decoder for incoming data
//	OnPeer        - Optional callback for handling new peers
type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

// TCPTransport manages TCP connections and message passing between peers.
// It implements the Transport interface for GoVaultFS.
type TCPTransport struct {
	TCPTransportOpts              // Configuration options
	listener         net.Listener // TCP listener for incoming connections
	rpcch            chan RPC     // Channel for incoming RPC messages
}

// NewTCPTransport creates a new TCPTransport with the given options.
// The rpcch channel buffers incoming messages for consumption.
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC, 1024),
	}
}

// Addr returns the address the transport is listening on (Transport interface).
func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

// Consume returns a read-only channel for incoming RPC messages (Transport interface).
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Close shuts down the TCP listener (Transport interface).
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial connects to a remote peer at the given address and starts handling the connection (Transport interface).
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	// Handle the new outbound connection in a separate goroutine
	go t.handleConn(conn, true)

	return nil
}

// ListenAndAccept starts the TCP listener and begins accepting incoming connections.
// It launches the accept loop in a goroutine and logs the listening address.
func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("TCP transport listening on port: %s\n", t.ListenAddr)

	return nil
}

// startAcceptLoop continuously accepts new incoming TCP connections.
// For each accepted connection, it launches handleConn in a goroutine.
func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return // Listener closed, exit loop
		}

		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}

		// Handle the new inbound connection in a separate goroutine
		go t.handleConn(conn, false)
	}
}

// handleConn manages a single peer connection, performing handshake, peer callback, and message read loop.
//   - If handshake fails, the connection is dropped.
//   - If OnPeer callback is set and fails, the connection is dropped.
//   - In the read loop, decodes incoming RPC messages and handles stream synchronization.
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	// Run handshake logic (e.g., authentication, protocol negotiation)
	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	// Optional callback for custom peer handling
	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop: decode messages and handle streams
	for {
		rpc := RPC{}
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return // On decode error, drop connection
		}

		rpc.From = conn.RemoteAddr().String() // Set sender address

		if rpc.Stream {
			// If this is a stream message, block until stream is closed
			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}

		// Forward regular RPC message to channel for consumption
		t.rpcch <- rpc
	}
}
