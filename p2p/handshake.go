// Handshake utilities for P2P connections in GoVaultFS
// This file defines the handshake function type and a no-op implementation for peer connections.
package p2p

// HandshakeFunc defines the signature for a handshake function between peers.
// It allows custom logic to be executed when establishing a connection with a peer.
// For example, authentication, protocol negotiation, or capability exchange.
type HandshakeFunc func(Peer) error

// NOPHandshakeFunc is a no-operation handshake function.
// It performs no handshake logic and always returns nil (success).
// Useful as a default or placeholder when no handshake is required.
func NOPHandshakeFunc(Peer) error { return nil }
