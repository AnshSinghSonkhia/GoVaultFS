# Distributed File System in Go

A comprehensive implementation of a decentralized, peer-to-peer (P2P) distributed file storage system built in Go. This project demonstrates building distributed systems from scratch, covering advanced topics in network programming, cryptography, and system design.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
  - [Peer-to-Peer Transport Layer](#peer-to-peer-transport-layer)
  - [Content-Addressable Storage](#content-addressable-storage)
  - [File Server Architecture](#file-server-architecture)
  - [Encryption and Security](#encryption-and-security)
- [How It Works](#how-it-works)
  - [File Storage Process](#file-storage-process)
  - [File Retrieval Process](#file-retrieval-process)
<!-- - [Installation](#installation) -->
- [Usage](#usage)
- [Technical Implementation](#technical-implementation)
- [Learning Outcomes](#learning-outcomes)
- [Real-World Applications](#real-world-applications)
- [Contributing](#contributing)
- [License](#license)

## Overview

This is a content-addressable storage (CAS) system that implements a fully distributed architecture where files are stored across multiple nodes in a peer-to-peer network. Unlike traditional file systems that rely on hierarchical paths or centralized servers, this system identifies and retrieves files based on their content using cryptographic hashing.

## Features

### Core Features
- **True P2P Architecture**: No central server or single point of failure
- **Content-Addressable Storage**: Files identified by their cryptographic hash
- **Automatic Deduplication**: Identical files share the same storage location
- **Encrypted Storage**: AES encryption for secure file storage and transmission
- **Streaming Support**: Efficient handling of large files
- **Fault Tolerance**: File replication across multiple nodes
- **Custom Network Protocol**: Built from scratch TCP transport layer

### Technical Features
- **Concurrent Operations**: Parallel file operations and network communications
- **Dynamic Peer Discovery**: Bootstrap nodes help new peers join the network
- **Data Integrity**: Cryptographic hashes ensure file integrity
- **Efficient Network Usage**: Content-based addressing minimizes duplicate transfers

## Architecture

### Peer-to-Peer Transport Layer

The system implements a custom TCP transport layer in the `p2p` package:

```go
// Transport Interface defines how nodes communicate
type Transport interface {
    Dial(string) error
    ListenAndAccept() error
    Consume() <-chan RPC
    Close() error
    Addr() string
}
```

**Key Components:**
- **Transport Interface**: Defines the contract for node communication
- **TCP Transport**: Handles TCP connections, handshakes, and message encoding/decoding
- **Peer Management**: Manages connections between nodes in the network
- **Message Protocol**: Custom protocol for efficient peer communication

### Content-Addressable Storage

The storage system uses CAS principles for file management:

```go
// Content-based file identification
type PathKey struct {
    PathName string
    FileName string
}
```

**Features:**
- **Content Hashing**: Files are hashed to create unique identifiers
- **Path Transformation**: Hierarchical directory structure based on content hashes (similar to Git)
- **Automatic Deduplication**: Identical files automatically share storage
- **Efficient Retrieval**: O(1) lookup time for file access

### File Server Architecture

The `FileServer` component orchestrates the entire system:

```go
type FileServer struct {
    transport p2p.Transport
    store     *Store
    peers     map[string]p2p.Peer
}
```

**Responsibilities:**
- **File Storage**: Local file operations and network distribution
- **File Retrieval**: Fetches files from local storage or network
- **Broadcasting**: Distributes files to connected peers
- **Peer Discovery**: Manages bootstrap nodes and connections

### Encryption and Security

Comprehensive security implementation using AES encryption:

```go
// Per-file encryption with unique keys
type EncryptionKey [32]byte

func (s *FileServer) StoreData(key string, r io.Reader) error {
    // Encrypt and store file
}
```

**Security Features:**
- **Per-File Encryption Keys**: Each file encrypted with unique key
- **Stream Encryption**: On-the-fly encryption/decryption during transmission
- **Data Integrity**: Cryptographic hashes verify file authenticity
- **Secure Communication**: Encrypted peer-to-peer transfers

## How It Works

### File Storage Process

1. **Content Hashing**: File is processed through hash function to generate unique identifier
   ```go
   hash := sha256.Sum256(fileData)
   ```

2. **Path Generation**: Hash creates hierarchical directory structure
   ```
   hash: a1b2c3d4e5f6...
   path: /a1/b2/c3/d4e5f6...
   ```

3. **Local Storage**: File is encrypted and stored using content-based path

4. **Network Distribution**: File broadcast to connected peers for redundancy

5. **Peer Synchronization**: Connected nodes receive and store file copies

### File Retrieval Process

1. **Local Check**: First attempts retrieval from local storage
   ```go
   if file, err := store.Get(key); err == nil {
       return file
   }
   ```

2. **Network Query**: Broadcasts request to network if not found locally

3. **Peer Response**: Nodes with file respond with data

4. **Stream Transfer**: Encrypted stream transfer between peers

5. **Local Caching**: Retrieved files cached for future access

## Usage

### Starting a Bootstrap Node
```bash
# Start the first node (bootstrap node)
./dfs --port=3000 --bootstrap
```

### Joining the Network
```bash
# Start additional nodes
./dfs --port=3001 --bootstrap-addr=localhost:3000
./dfs --port=3002 --bootstrap-addr=localhost:3000
```

### Storing Files
```go
// Example: Store a file in the network
server := NewFileServer(opts)
err := server.StoreFile("document.pdf", fileReader)
```

### Retrieving Files
```go
// Example: Retrieve a file from the network
reader, err := server.GetFile(fileHash)
if err != nil {
    log.Fatal(err)
}
defer reader.Close()
```

## Technical Implementation

### Advanced Go Concepts Demonstrated

**Interface-Based Design**
```go
type Store interface {
    Put(string, io.Reader) error
    Get(string) (io.ReadCloser, error)
    Has(string) bool
    Delete(string) error
}
```

**Concurrent Programming**
- Goroutines for parallel operations
- Channels for communication
- Sync primitives for thread safety

**Stream Processing**
- Efficient I/O with readers/writers
- Minimal memory footprint for large files

**Error Handling**
- Comprehensive error propagation
- Context-aware error messages

## Learning Outcomes

This project provides hands-on experience with:

- **Distributed Systems Design**: Building systems that operate across multiple machines
- **Network Programming**: TCP programming, custom protocols, and P2P communication
- **Cryptography**: Practical implementation of encryption, hashing, and data integrity
- **System Architecture**: Designing scalable, fault-tolerant systems
- **Content-Addressable Storage**: Principles behind modern distributed storage
- **Go Best Practices**: Interface design, concurrency, and error handling

## Real-World Applications

The concepts implemented here are used in production systems:

- **IPFS (InterPlanetary File System)**: Similar content-addressable storage principles
- **Git Version Control**: Content-based addressing for version management
- **Blockchain Systems**: Content-addressable storage for data integrity
- **Content Delivery Networks**: Distributed storage approaches
- **Distributed Databases**: Peer replication and consistency

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Guidelines
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Development

### Windows Setup

This project uses a Makefile for build automation. On Windows, you'll need to install GNU Make:

```powershell
winget install GnuWin32.Make
```

After installation, run the setup script to add make to your PATH:

```powershell
.\setup-make.ps1
```

### Available Commands

- `make build` - Build the application
- `make run` - Build and run the application
- `make test` - Run tests

### Alternative (PowerShell Scripts)

If you prefer not to use make, PowerShell scripts are also available:

- `.\build.ps1` - Build the application
- `.\run.ps1` - Build and run the application
- `.\test.ps1` - Run tests