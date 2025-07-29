// This file demonstrates the initialization and operation of a distributed, peer-to-peer file system.
// It sets up three file server nodes, connects them, and runs a test scenario to store, delete, and retrieve files across the network.
package main


import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/AnshSinghSonkhia/GoVaultFS/p2p"
)


// makeServer creates and configures a new FileServer node.
// listenAddr: TCP address to listen on (e.g., ":3000")
// nodes: addresses of bootstrap peers to connect to
func makeServer(listenAddr string, nodes ...string) *FileServer {
	// Configure TCP transport layer for P2P communication
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc, // No-op handshake for demo
		Decoder:       p2p.DefaultDecoder{}, // Default message decoder
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	// Windows compatibility: replace ':' in port with 'port' for valid directory names
	storageRoot := strings.ReplaceAll(listenAddr, ":", "port") + "_network"

	// Configure file server options
	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),      // Generate a new AES encryption key
		StorageRoot:       storageRoot,             // Local storage directory
		PathTransformFunc: CASPathTransformFunc,    // Hash-to-path converter
		Transport:         tcpTransport,            // Network transport layer
		BootstrapNodes:    nodes,                   // List of bootstrap peers
	}

	// Create the FileServer instance
	s := NewFileServer(fileServerOpts)

	// Set up peer connection handler
	tcpTransport.OnPeer = s.OnPeer

	return s
}


// main demonstrates the distributed file system in action.
// It sets up three file server nodes, connects them, and runs a test scenario.
// The time.Sleep calls in the code are used to introduce delays between the startup
// of the services (s1, s2, and s3) to ensure proper sequencing and stabilization
// of the system. Below is a breakdown of their purpose:

// 1. time.Sleep(500 * time.Millisecond)
//    - This delay is introduced after starting s1 and before starting s2.
//    - The purpose is to give s1 enough time to initialize and start running before
//      s2 attempts to start. If s2 depends on s1 being fully operational (e.g., for
//      network connections or shared resources), this delay ensures that s1 is ready.

// 2. time.Sleep(2 * time.Second) (after starting s2)
//    - This delay is used to allow the network or system to stabilize after both
//      s1 and s2 have started.
//    - If s1 and s2 need to establish communication or synchronize with each other,
//      this delay ensures that they have enough time to complete those operations
//      before s3 starts.

// 3. time.Sleep(2 * time.Second) (after starting s3)
//    - This delay is used to allow s3 to fully initialize and connect to s1 and s2.
//    - If s3 depends on s1 and s2 being fully operational and connected, this delay
//      ensures that s3 has enough time to stabilize before the program continues.

// Why is this necessary?
// - Concurrency Issues: Since the services are started in separate goroutines, they
//   run concurrently. Without these delays, there’s no guarantee that one service
//   will be ready before another starts interacting with it.
// - Initialization Dependencies: If s2 or s3 depend on s1 being fully initialized,
//   starting them too early could lead to errors or undefined behavior.
// - Network Stabilization: In distributed systems, it often takes time for nodes to
//   establish connections, synchronize, or stabilize. These delays simulate that
//   waiting period.
func main() {
	// Create three file server nodes:
	// s1: listens on :3000 (standalone)
	// s2: listens on :7000 (standalone)
	// s3: listens on :5000, bootstraps to :3000 and :7000 
	s1 := makeServer(":3000", "")
	s2 := makeServer(":7000", "")
	s3 := makeServer(":5000", ":3000", ":7000")

	// Start s1 and s2 in background goroutines
	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(500 * time.Millisecond) // Allow s1 to start before s2
	go func() { log.Fatal(s2.Start()) }()

	// Wait for network stabilization
	time.Sleep(2 * time.Second)

	// Start s3 (connects to s1 and s2)
	go s3.Start()
	time.Sleep(2 * time.Second)

	// Test scenario: Store, delete, and retrieve 20 files
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("picture_%d.png", i) // Unique file key
		data := bytes.NewReader([]byte("my big data file here!")) // File content

		// Store file on s3 (will be replicated to peers)
		s3.Store(key, data)

		// Delete local copy to force network retrieval
		if err := s3.store.Delete(s3.ID, key); err != nil {
			log.Fatal(err)
		}

		// Retrieve file from network (should fetch from peers)
		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		// Read and print file content to verify successful retrieval
		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))
	}
}
