// Cryptography utilities for GoVaultFS
// This file provides functions for generating IDs, hashing keys, and encrypting/decrypting file streams.
// All encryption uses AES in CTR mode for secure, efficient file storage and transfer.

// CTR (Counter) Mode is a block cipher mode of operation for symmetric encryption algorithms like AES. In CTR mode, a unique "counter" value (often combined with an initialization vector, IV) is encrypted for each block, and the result is XORed with the plaintext to produce ciphertext (or vice versa for decryption).

// - Allows parallel encryption/decryption of blocks.
// - Turns a block cipher into a stream cipher.
// - The counter/IV must be unique for each encryption to ensure security.
// - Used for efficient, random-access encryption of data streams.
// So here, AES-CTR mode is used for encrypting and decrypting file streams securely and efficiently.

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"io"
)

// generateID creates a random 32-byte hex string for node or file identification
func generateID() string {
	buf := make([]byte, 32)
	io.ReadFull(rand.Reader, buf)
	return hex.EncodeToString(buf)
}

// hashKey returns an MD5 hash of the given key as a hex string
// Used for content-addressable storage and network lookup
func hashKey(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

// newEncryptionKey generates a new random 32-byte AES key
// Each node uses its own key for file encryption
func newEncryptionKey() []byte {
	keyBuf := make([]byte, 32)
	io.ReadFull(rand.Reader, keyBuf)
	return keyBuf
}

// copyStream encrypts or decrypts data from src to dst using the given cipher stream
// Used for both encryption and decryption in CTR mode
func copyStream(stream cipher.Stream, blockSize int, src io.Reader, dst io.Writer) (int, error) {
	var (
		buf = make([]byte, 32*1024) // 32KB buffer for efficient streaming
		nw  = blockSize             // Track total bytes written
	)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf, buf[:n]) // Encrypt/decrypt in-place
			nn, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			nw += nn
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return nw, nil
}

// copyDecrypt decrypts data from src to dst using AES-CTR mode
// Reads the IV from the beginning of src, then streams decryption
func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	// Read IV (initialization vector) from src
	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}

	stream := cipher.NewCTR(block, iv)
	return copyStream(stream, block.BlockSize(), src, dst)
}

// copyEncrypt encrypts data from src to dst using AES-CTR mode
// Generates a random IV, prepends it to dst, then streams encryption
func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize()) // 16 bytes for AES
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	// Prepend IV to the output file/stream
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}

	stream := cipher.NewCTR(block, iv)
	return copyStream(stream, block.BlockSize(), src, dst)
}
