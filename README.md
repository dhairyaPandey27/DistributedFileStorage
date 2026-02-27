# PeerVault - Distributed File Storage System

A peer-to-peer distributed file storage system built in Go that enables decentralized file storage and retrieval across a network of nodes. Files are automatically encrypted and replicated across multiple peers for redundancy and availability.

## Features

### Core Functionality
- Distributed File Storage: Store files across multiple peer nodes in a decentralized network
- File Retrieval: Retrieve files from any node in the network, with automatic discovery and loading from local or remote storage
- File Deletion: Delete files across all nodes in the network with a single command
- Auto-Replication: Files are automatically replicated and distributed to all connected peers for redundancy

### Security & Encryption
- AES Encryption: Secure file encryption using AES-256 in CTR (Counter) mode
- Random Key Generation: Each server generates its own encryption key for file security
- Hashed Key Storage: File keys are hashed using MD5 to create unique identifiers
- Secure Content-Addressed Storage: SHA1-based content addressing for efficient and secure file location

### Networking
- TCP-Based P2P Communication: Peer-to-peer communication over TCP with custom message protocol
- Bootstrap Nodes: Support for joining existing networks through bootstrap node addresses
- Dynamic Peer Discovery: Automatic discovery and connection management with peers
- RPC-Style Messaging: Gob-encoded message serialization for inter-node communication
- Concurrent Operations: Full support for concurrent file operations across the network

### Architecture
- Message Protocol: Custom RPC protocol with support for MessageStoreFile (store file), MessageGetFile (request file), and MessageDeleteFile (delete file from all nodes), with efficient stream and message types

- Content-Addressed File System: Uses CAS (Content-Addressed Storage) PathTransform function for efficient file organization with SHA1 hashing, deterministic file paths based on content, and scalable storage directory structure

## How It Works

### Node Architecture
Each DistShare node operates as both a server and a client:
1. **Listens** for incoming connections from other peers
2. **Maintains** a local encrypted file store
3. **Broadcasts** file operations to all connected peers
4. **Replicates** files from other peers automatically

### File Storage Process
When a file is stored:
1. The local node encrypts and saves the file to disk
2. A broadcast message is sent to all peers with file metadata
3. Peers receive the file stream and decrypt it locally
4. File is now available across the entire network

### File Retrieval Process
When a file is requested:
1. Check local storage first for immediate retrieval
2. If not found locally, broadcast a file request to all peers
3. Peers respond by streaming their copy of the file
4. Remote node decrypts and stores the file locally
5. File is now cached locally for future access

### Network Join
New nodes can join the network by:
1. Providing bootstrap node addresses on startup
2. Establishing TCP connections to bootstrap nodes
3. Receiving peer information from connected nodes
4. Building connections with all known peers

## Try It On Your System

### Clone the Repository
```bash
git clone https://github.com/dhairyaPandey27/PeerVault.git
cd PeerVault
```

### Run the Distributed Network
```bash
make run
```

This command will start a 3-node distributed file storage network:
- Server 1: Listening on port 3000 (standalone bootstrap node)
- Server 2: Listening on port 7000 (standalone bootstrap node)
- Server 3: Listening on port 8000 (joins servers 1 and 2)

The demo will then:
1. Store a test file across the network
2. Wait for replication to complete
3. Delete the file from all nodes
4. Retrieve and display the file content

### Verify the Build
```bash
make build
```

This will compile the project and create the executable in the bin/ directory.

## Technical Stack

- **Language**: Go 1.24.0
- **Cryptography**: crypto/aes, crypto/cipher (AES-256-CTR)
- **Serialization**: Go's gob encoding
- **Networking**: TCP with custom protocol
- **Concurrency**: Goroutines and sync.Mutex for thread-safe operations

## Project Structure

```
├── main.go                      # Entry point with demo implementation
├── server.go                    # FileServer implementation with message handling
├── store.go                     # Local file storage and encryption handling
├── crypto.go                    # Encryption, decryption, and ID generation
├── p2p/
│   ├── transport.go             # Transport interface definition
│   ├── tcp_transport.go         # TCP implementation of Transport
│   ├── encoding.go              # RPC message encoding/decoding
│   ├── handshake.go             # Peer handshake protocol
│   ├── message.go               # RPC message types
│   ├── tcp_transport_test.go    # Transport tests
│   └── encoding.go              # Encoder tests
├── go.mod                       # Go module definition
├── Makefile                     # Build and run commands
└── bin/                         # Compiled executable output
```

## Design Decisions

### Encryption Strategy
Files are encrypted before being written to disk or transmitted over the network. Each node generates its own encryption key, and encryption happens at both:
- **Storage layer**: When writing files to local disk
- **Network layer**: When transmitting files between peers

### Content-Addressed Storage
Using SHA1-based content addressing provides:
- Deterministic file locations
- Efficient deduplication potential
- Quick verification of file integrity
- Scalable directory hierarchies

### Broadcast Replication
All file operations are broadcast to maintain eventual consistency:
- High availability through full replication
- Automatic failover (can retrieve from any peer)
- Simple and reliable consistency model
- Suitable for moderately sized networks

## Future Enhancements

Potential improvements for production use:
- Configurable replication factor instead of full replication
- Quorum-based consistency for distributed consensus
- Peer discovery via DHT (Distributed Hash Table)
- Partial file synchronization for large files
- REST API for external access
- Web dashboard for monitoring
- Load balancing and peer selection optimization

---

**Project**: PeerVault - Distributed File Storage  
**Author**: Dhairya Pandey  
**License**: MIT
