// Unit test for encryption and decryption functions in GoVaultFS
// This test verifies that data encrypted with copyEncrypt can be correctly decrypted with copyDecrypt using AES-CTR mode.
package main

import (
	"bytes"
	"fmt"
	"testing"
)

// TestCopyEncryptDecrypt checks that encryption and decryption work as expected.
// It encrypts a string, then decrypts it, and verifies the output matches the original.
func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "Foo not bar"                // Test data to encrypt
	src := bytes.NewReader([]byte(payload)) // Source reader for encryption
	dst := new(bytes.Buffer)                // Destination buffer for encrypted data
	key := newEncryptionKey()               // Generate a random AES key

	// Encrypt the payload using AES-CTR
	_, err := copyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	// Print lengths for debugging (optional)
	fmt.Println(len(payload))      // Length of original data
	fmt.Println(len(dst.String())) // Length of encrypted data (includes IV)

	out := new(bytes.Buffer) // Buffer for decrypted output
	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	// The decrypted output should match the original payload
	// nw should be 16 (IV) + payload length
	if nw != 16+len(payload) {
		t.Fail()
	}

	if out.String() != payload {
		t.Errorf("decryption failed!!!")
	}
}
