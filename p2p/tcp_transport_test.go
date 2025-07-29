// Unit tests for TCPTransport in GoVaultFS
// This file verifies basic initialization and listening behavior of the TCP transport layer.
package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTCPTransport checks that a TCPTransport can be created and started with default options.
// It verifies:
//   - The transport is initialized with the correct listen address
//   - The transport can start listening and accepting connections without error
func TestTCPTransport(t *testing.T) {
	// Setup transport options with a test address, no-op handshake, and default decoder
	opts := TCPTransportOpts{
		ListenAddr:    ":3000",          // Listen on port 3000 for test
		HandshakeFunc: NOPHandshakeFunc, // No handshake logic for simplicity
		Decoder:       DefaultDecoder{}, // Use default decoder for messages
	}
	// Create a new TCPTransport instance
	tr := NewTCPTransport(opts)
	// Ensure the transport's listen address is set correctly
	assert.Equal(t, tr.ListenAddr, ":3000")

	// Attempt to start listening and accepting connections; should not return an error
	assert.Nil(t, tr.ListenAndAccept())
}
