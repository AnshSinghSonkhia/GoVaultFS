// Message definitions for P2P communication in GoVaultFS
// This file provides constants and the RPC struct for network messaging between nodes.
package p2p

// Message type constants used to identify the kind of network message received.
const (
	IncomingMessage = 0x1 // Indicates a regular message with payload
	IncomingStream  = 0x2 // Indicates a stream message (e.g., file transfer)
)

// RPC represents a Remote Procedure Call message sent between nodes.
// It is the main data structure for exchanging information over the transport layer.
// Fields:
//   From    - The sender's node ID or address
//   Payload - The actual message data or file chunk
//   Stream  - True if this message is part of a stream (e.g., file transfer)
type RPC struct {
	From    string // Sender identifier
	Payload []byte // Message or file data
	Stream  bool   // Stream flag for file/data streaming
}
