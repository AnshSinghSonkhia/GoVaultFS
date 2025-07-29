// Transport and Peer interfaces for GoVaultFS P2P networking
// This file defines abstractions for remote nodes and communication channels in the network.
package p2p

import "net"

// Peer abstracts a remote node in the network.
// It embeds net.Conn for low-level network operations and adds methods for sending data and managing streams.
//   Send([]byte) error   - Send raw bytes to the peer
//   CloseStream()        - Signal the end of a stream (e.g., file transfer)
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Transport abstracts any communication channel between nodes (TCP, UDP, WebSockets, etc).
// It provides methods for connection management and message consumption:
//   Addr() string             - Get the listening address
//   Dial(string) error        - Connect to a remote node
//   ListenAndAccept() error   - Start listening and accepting connections
//   Consume() <-chan RPC      - Read-only channel for incoming RPC messages
//   Close() error             - Shut down the transport
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
