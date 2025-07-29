// Content-addressable storage implementation for GoVaultFS
// This file provides the logic for storing, retrieving, and managing files using their content hash.
// Files are organized in a hierarchical directory structure based on their hash for efficient deduplication and lookup.
package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// Default root folder for all file storage
const defaultRootFolderName = "ggnetwork"

// CASPathTransformFunc transforms a file key into a hierarchical path using SHA-1 hash
// This enables content-addressable storage and deduplication
func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blocksize := 5 // Split hash into blocks for directory structure
	sliceLen := len(hashStr) / blocksize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"), // e.g. "d44bb/d0bbd/a685d/..."
		Filename: hashStr,                  // Full hash as filename
	}
}

// PathTransformFunc is a function type for converting keys to path structures
// Allows pluggable path transformation logic
type PathTransformFunc func(string) PathKey

// PathKey holds the directory path and filename for a file
type PathKey struct {
	PathName string // Hierarchical directory path
	Filename string // Full hash filename
}

// FirstPathName returns the first directory in the path
func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

// FullPath returns the full path including directories and filename
func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

// StoreOpts holds configuration for a Store instance
type StoreOpts struct {
	Root              string            // Root directory for all files
	PathTransformFunc PathTransformFunc // Function to transform keys to paths
}

// DefaultPathTransformFunc is a fallback path transformer (no hashing)
var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

// Store manages file storage and retrieval on disk
type Store struct {
	StoreOpts
}

// NewStore creates a new Store with the given options
func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}

	return &Store{
		StoreOpts: opts,
	}
}

// Has checks if a file exists for the given node ID and key
func (s *Store) Has(id string, key string) bool {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

// Clear deletes all files and directories under the root
func (s *Store) Clear() error {
	return os.RemoveAll(s.Root)
}

// Delete removes the file and its directory for the given node ID and key
func (s *Store) Delete(id string, key string) error {
	pathKey := s.PathTransformFunc(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.Filename)
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FirstPathName())

	return os.RemoveAll(firstPathNameWithRoot)
}

// Write saves a file stream to disk for the given node ID and key
func (s *Store) Write(id string, key string, r io.Reader) (int64, error) {
	return s.writeStream(id, key, r)
}

// WriteDecrypt decrypts and writes an encrypted file stream to disk
func (s *Store) WriteDecrypt(encKey []byte, id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	defer f.Close() // Ensure file is closed after writing
	n, err := copyDecrypt(encKey, r, f)
	return int64(n), err
}

// openFileForWriting creates all necessary directories and opens the file for writing
func (s *Store) openFileForWriting(id string, key string) (*os.File, error) {
	pathKey := s.PathTransformFunc(key)
	pathNameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.PathName)
	if err := os.MkdirAll(pathNameWithRoot, os.ModePerm); err != nil {
		return nil, err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	return os.Create(fullPathWithRoot)
}

// writeStream writes a file stream to disk, ensuring the file is closed after writing
func (s *Store) writeStream(id string, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(id, key)
	if err != nil {
		return 0, err
	}
	defer f.Close() // Ensure file is closed after writing
	return io.Copy(f, r)
}

// Read returns a file stream and its size for the given node ID and key
func (s *Store) Read(id string, key string) (int64, io.Reader, error) {
	return s.readStream(id, key)
}

// readStream opens a file for reading and returns its size and stream
func (s *Store) readStream(id string, key string) (int64, io.ReadCloser, error) {
	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, id, pathKey.FullPath())

	file, err := os.Open(fullPathWithRoot)
	if err != nil {
		return 0, nil, err
	}

	fi, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}

	return fi.Size(), file, nil
}
