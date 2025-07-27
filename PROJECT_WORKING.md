# GoVaultFS Project Documentation

## Project Overview
GoVaultFS is a **distributed peer-to-peer (P2P) file storage system** built in Go that implements content-addressable storage (CAS) with encryption. It creates a decentralized network where files are stored across multiple nodes without a central server, identified by their cryptographic hash rather than traditional file paths.

## Project Structure
```
GoVaultFS/
├── main.go                  # Application entry point and demo
├── server.go               # Core file server implementation
├── store.go                # File storage and retrieval logic
├── crypto.go               # Encryption/decryption utilities
├── p2p/                    # Peer-to-peer networking layer
│   ├── transport.go        # Transport interface definitions
│   ├── tcp_transport.go    # TCP transport implementation
│   ├── handshake.go        # Peer handshake protocol
│   ├── message.go          # Message types and structures
│   └── encoding.go         # Data encoding/decoding
├── bin/                    # Compiled binaries
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── Makefile               # Build automation
├── README.md              # Project documentation
└── *_test.go              # Test files

```

## Key Features
- **True P2P Architecture**: No central server, fully distributed
- **Content-Addressable Storage**: Files identified by SHA-1 hash
- **Automatic Deduplication**: Identical files share storage space
- **AES Encryption**: Secure file storage and transmission
- **Fault Tolerance**: File replication across multiple nodes
- **Custom TCP Protocol**: Built-from-scratch networking layer

## Core Components

### P2P Transport Layer (`p2p/`)
- **TCP Transport**: Custom TCP-based communication protocol
- **Peer Management**: Connection handling and peer discovery
- **Message Encoding**: Binary message serialization using GOB
- **Handshake Protocol**: Secure peer authentication and connection setup

### File Server (`server.go`)
- **Distributed Storage**: Manages file storage across network nodes
- **Peer Coordination**: Handles communication between network peers
- **File Replication**: Ensures files are replicated across multiple nodes
- **Network Bootstrap**: Connects to existing nodes to join the network

### Storage System (`store.go`)
- **Content-Addressable Storage (CAS)**: Files identified by SHA-1 hash
- **Path Transformation**: Converts file keys to hierarchical directory structure
- **Local File Management**: Handles reading/writing files to disk
- **Deduplication**: Prevents storing duplicate files

### Cryptography (`crypto.go`)
- **AES Encryption**: File content encryption/decryption
- **Key Generation**: Secure random key generation for each node
- **Streaming Encryption**: Efficient encryption for large files
- **ID Generation**: Unique node identifier generation

## How The System Works

### 1. Network Initialization
- Each node starts with a unique ID and encryption key
- Nodes listen on specified TCP ports (e.g., :3000, :7000, :5000)
- Bootstrap nodes help new peers discover and join the network
- Peers maintain connections to multiple other nodes for redundancy

### 2. File Storage Process
1. **File Input**: Client provides a file with a key (filename)
2. **Hash Generation**: System generates SHA-1 hash of the key
3. **Path Creation**: Hash is split into directory structure (e.g., `d44bb/d0bbd/a685d/...`)
4. **Encryption**: File content is encrypted using node's AES key
5. **Local Storage**: File is stored locally using the hash-based path
6. **Network Replication**: File is replicated to connected peer nodes
7. **Verification**: Other nodes confirm successful storage

### 3. File Retrieval Process
1. **File Request**: Client requests file by key
2. **Local Check**: Node first checks if file exists locally
3. **Network Query**: If not found locally, queries connected peers
4. **File Transfer**: Peer nodes stream the file over TCP connection
5. **Decryption**: Received encrypted data is decrypted
6. **Local Caching**: Retrieved file is cached locally for future access

### 4. Content-Addressable Storage (CAS)
- Files are identified by their content hash, not filename
- Same content = same hash = no duplication
- Directory structure: `{nodeID}/{hash_part1}/{hash_part2}/.../{full_hash}`
- Example path: `port5000_network/d44bb/d0bbd/a685d/d44bbd0bbda685d5db90f419568b531ab9afa97b`

## Network Protocol

### Message Types
- **MessageStoreFile**: Requests to store file on remote node
- **MessageGetFile**: Requests to retrieve file from remote node
- **RPC (Remote Procedure Call)**: Communication wrapper for all messages

### Connection Flow
1. **TCP Connection**: Establish TCP connection between peers
2. **Handshake**: Exchange node information and capabilities
3. **Message Exchange**: Send/receive file storage and retrieval requests
4. **Stream Handling**: Manage concurrent file transfers
5. **Connection Cleanup**: Proper connection termination

## Technical Implementation

### Current Demo Implementation
The `main.go` demonstrates the system with:
- **3 File Servers**: Running on ports 3000, 7000, and 5000
- **Network Topology**: Port 5000 connects to both 3000 and 7000
- **Test Scenario**: Stores 20 test files, deletes them locally, then retrieves from network
- **File Operations**: Store → Delete → Get → Verify content

### Key Data Structures
```go
type FileServer struct {
    ID                string              // Unique node identifier
    EncKey            []byte              // AES encryption key
    StorageRoot       string              // Local storage directory
    PathTransformFunc PathTransformFunc   // Hash-to-path converter
    Transport         p2p.Transport       // Network transport layer
    BootstrapNodes    []string            // Known peer addresses
    peers             map[string]p2p.Peer // Connected peers
    store             *Store              // Local file storage
}

type Store struct {
    Root              string              // Root storage directory
    PathTransformFunc PathTransformFunc   // Path transformation function
}

type PathKey struct {
    PathName string // Directory path (e.g., "d44bb/d0bbd/a685d")
    Filename string // Full hash filename
}
```

## Build and Run
```bash
# Build the project
make build

# Run the application (starts 3-node demo)
make run

# Run tests
make test

# Clean build artifacts (not implemented)
make clean
```

## Current System Behavior
When you run `make run`, the system:

1. **Starts 3 File Servers**:
   - Server 1: Port 3000 (standalone)
   - Server 2: Port 7000 (standalone) 
   - Server 3: Port 5000 (connects to 3000 and 7000)

2. **Establishes Network**:
   - Servers start listening on their respective ports
   - Port 5000 connects to ports 3000 and 7000
   - TCP connections are established between peers

3. **Runs Test Scenario**:
   - Creates 20 test files named `picture_0.png` to `picture_19.png`
   - Each file contains the text "my big data file here!"
   - Files are stored across the network with encryption
   - Local copies are deleted to test network retrieval
   - Files are retrieved from peer nodes and verified

4. **Output Shows**:
   - Connection establishment logs
   - File storage confirmations ("written X bytes to disk")
   - File deletion confirmations
   - Network retrieval messages ("fetching from network...")
   - Content verification (prints file content)

## Dependencies
- **Go Standard Library**: Core networking, crypto, and I/O operations
- **github.com/stretchr/testify**: Testing framework for unit tests
- **No external frameworks**: Pure Go implementation

## Security Features
- **AES Encryption**: All files are encrypted before storage and network transmission
- **Unique Node Keys**: Each node has its own encryption key
- **Content Integrity**: SHA-1 hashes ensure file integrity
- **Secure Key Generation**: Cryptographically secure random key generation
- **Stream Encryption**: Large files are encrypted in chunks for efficiency

## Windows Compatibility Fixes
The project includes specific fixes for Windows:
- **Path Sanitization**: Replaces `:` in port numbers with `port` for valid Windows paths
- **File Handle Management**: Proper file closing to prevent "file in use" errors
- **Directory Creation**: Ensures all parent directories are created

## Development Status
This project demonstrates a working distributed file system with:
- ✅ Peer-to-peer networking layer
- ✅ Content-addressable storage
- ✅ File encryption/decryption
- ✅ Network file replication
- ✅ Automatic peer discovery
- ✅ Fault-tolerant file retrieval
- ✅ Windows compatibility

## Learning Outcomes
This project demonstrates:
- **Distributed Systems**: Understanding P2P networks and consensus
- **Network Programming**: TCP connections, protocol design, message handling
- **Cryptography**: Symmetric encryption, hashing, secure key management
- **File Systems**: Content-addressable storage, path transformation
- **Concurrency**: Goroutines, channels, concurrent file operations
- **System Design**: Fault tolerance, scalability, data replication

## Real-World Applications
Similar systems are used in:
- **Git Version Control**: Content-addressable object storage
- **IPFS (InterPlanetary File System)**: Distributed web infrastructure
- **BitTorrent**: Peer-to-peer file sharing
- **Blockchain Storage**: Decentralized data storage
- **CDN Systems**: Content distribution networks

## Potential Enhancements
- Web-based UI for file management
- RESTful API endpoints
- File metadata and versioning
- Advanced peer discovery mechanisms
- Load balancing and sharding
- Database integration for metadata
- Authentication and access control
- Network topology optimization