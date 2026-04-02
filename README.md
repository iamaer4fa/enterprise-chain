# Enterprise Chain

A lightweight, enterprise-grade Layer 1 blockchain built entirely from scratch in Go. 

Enterprise Chain is a custom-built, programmable state machine designed to demonstrate the core mechanics of distributed systems, cryptography, and consensus. It operates as a fully decentralized Peer-to-Peer (P2P) network, utilizing a Proof-of-Authority (PoA) consensus mechanism optimized for private or consortium enterprise environments.

**Author:** Allan Registos  
**Module:** `github.com/iamaer4a/enterprise-chain`

## 🧠 Core Architecture

* **Language:** Go (Golang)
* **State Management:** [BadgerDB](https://github.com/dgraph-io/badger) (Embedded, lightning-fast Key-Value store)
* **Consensus:** Proof-of-Authority (PoA) with a fixed-interval Validator Minter loop.
* **Cryptography:** ECDSA (Elliptic Curve Digital Signature Algorithm) using the P-256 curve for private/public key generation, transaction signing, and verification.
* **Networking:** Custom TCP-based P2P Gossip Protocol for mempool synchronization and block propagation.

## ✨ Key Features

### 1. Account-Based State Model
Unlike UTXO-based chains (like Bitcoin), Enterprise Chain uses an Account-based model (similar to Ethereum). It features mathematically enforced replay protection via account nonces and tracks complete IN/OUT transaction histories for every address.

### 2. Native Smart Contracts (Chaincode)
The network supports native chaincode execution. Transactions can carry arbitrary JSON data payloads which are intercepted by the Virtual Machine during the `ProcessBlock` execution phase. It currently features a Decentralized Key-Value Registry (similar to ENS), allowing users to permanently bind data to smart contract memory addresses.

### 3. Integrated CLI Wallet
The node binary doubles as a cryptographic wallet and network client. 
* Generate secure ECDSA keypairs.
* Construct, sign, and broadcast transactions directly from the terminal.

### 4. RESTful API & Block Explorer
The node exposes a local HTTP API to interact with the decentralized ledger. The repository includes a lightweight, zero-dependency HTML/JS frontend Block Explorer that queries the network in real-time to visualize block height, mempool status, account balances, transaction histories, and smart contract memory states.

## 🚀 Getting Started

### Prerequisites
* Go 1.20+

### Building the Node
```bash
go mod tidy
go build -o enterprise-node