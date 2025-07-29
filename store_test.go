// Unit tests for the Store (content-addressable storage) in GoVaultFS
// These tests verify path transformation, file writing, reading, existence checks, and deletion.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
)

// TestPathTransformFunc checks that CASPathTransformFunc correctly transforms a key
// into the expected hierarchical path and filename using SHA-1 hashing.
// This ensures content-addressable storage works as designed.
func TestPathTransformFunc(t *testing.T) {
	key := "momsbestpicture"
	pathKey := CASPathTransformFunc(key)
	expectedFilename := "6804429f74181a63c50c3d81d733a12f14a353ff"
	expectedPathName := "68044/29f74/181a6/3c50c/3d81d/733a1/2f14a/353ff"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}

	if pathKey.Filename != expectedFilename {
		t.Errorf("have %s want %s", pathKey.Filename, expectedFilename)
	}
}

// TestStore validates the Store's core functionality:
// writing, reading, checking existence, and deleting files.
// It runs multiple iterations to ensure reliability and correctness.
func TestStore(t *testing.T) {
	s := newStore()
	id := generateID()
	defer teardown(t, s)

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("foo_%d", i)
		data := []byte("some jpg bytes")

		// Write data to the store
		if _, err := s.writeStream(id, key, bytes.NewReader(data)); err != nil {
			t.Error(err)
		}

		// Check if the file exists
		if ok := s.Has(id, key); !ok {
			t.Errorf("expected to have key %s", key)
		}

		// Read the file and verify its contents
		_, r, err := s.Read(id, key)
		if err != nil {
			t.Error(err)
		}

		b, _ := ioutil.ReadAll(r)
		if string(b) != string(data) {
			t.Errorf("want %s have %s", data, b)
		}

		// Delete the file
		if err := s.Delete(id, key); err != nil {
			t.Error(err)
		}

		// Ensure the file no longer exists
		if ok := s.Has(id, key); ok {
			t.Errorf("expected to NOT have key %s", key)
		}
	}
}

// newStore creates a new Store instance with CAS path transformation.
// Used for test setup.
func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

// teardown cleans up the store after each test by clearing all files.
// Ensures test isolation and prevents leftover data between tests.
func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
