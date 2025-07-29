// Encoding and decoding utilities for P2P network messages in GoVaultFS
// This file provides decoders for handling both structured (GOB) and raw stream messages over the network.
package p2p

import (
	"encoding/gob"
	"io"
)

// Decoder is an interface for decoding network messages into RPC structs
type Decoder interface {
	Decode(io.Reader, *RPC) error
}

// GOBDecoder decodes structured messages using Go's gob encoding
type GOBDecoder struct{}

// Decode decodes a gob-encoded message from the reader into the RPC struct
func (dec GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

// DefaultDecoder handles both stream and raw byte messages
// Used for decoding simple network signals and payloads
type DefaultDecoder struct{}

// Decode reads the first byte to check for a stream signal.
// If it's a stream, sets msg.Stream and returns.
// Otherwise, reads up to 1028 bytes as the message payload.
func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC) error {
	peekBuf := make([]byte, 1)
	if _, err := r.Read(peekBuf); err != nil {
		return nil // If nothing to read, just return
	}

	// If the first byte is IncomingStream, this is a raw stream (not a structured message)
	// We set Stream=true so the rest of the system can handle it appropriately
	stream := peekBuf[0] == IncomingStream
	if stream {
		msg.Stream = true
		return nil
	}

	// Otherwise, read the next 1028 bytes as the payload
	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	msg.Payload = buf[:n]

	return nil
}
