package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iamaer4a/enterprise-chain/api"
	"github.com/iamaer4a/enterprise-chain/core"
	"github.com/iamaer4a/enterprise-chain/core/consensus"
	"github.com/iamaer4a/enterprise-chain/core/crypto"
	"github.com/iamaer4a/enterprise-chain/core/mempool"
	"github.com/iamaer4a/enterprise-chain/core/types"
	"github.com/iamaer4a/enterprise-chain/database"
	"github.com/iamaer4a/enterprise-chain/network"
)

func main() {
	// Define CLI Flags
	dataStr := flag.String("data", "", "Smart contract JSON payload")
	nodeMode := flag.Bool("node", false, "Start the blockchain node")
	genKeyMode := flag.Bool("genkey", false, "Generate a new ECDSA keypair")
	dbPath := flag.String("db", "./chaindata", "Path to BadgerDB storage")
	p2pAddr := flag.String("p2p", ":3000", "P2P network listen address")
	apiAddr := flag.String("api", ":8080", "REST API listen address")

	// NEW: Wallet Flags
	txMode := flag.Bool("tx", false, "Send a transaction")
	privKeyHex := flag.String("privkey", "", "Sender's private key (hex)")
	toAddr := flag.String("to", "", "Recipient address (hex)")
	amountStr := flag.String("amount", "0", "Amount to send")
	nonce := flag.Uint64("nonce", 0, "Sender's account nonce")

	// NEW: Peer dialing flag
	peerAddr := flag.String("peer", "", "Address of a peer to connect to (e.g., localhost:3000)")

	flag.Parse()

	// Handle Key Generation CLI tool
	if *genKeyMode {
		privKey, _ := crypto.GenerateKey()

		// Calculate the address from the public key
		var address types.Address
		copy(address[:], privKey.PublicKey.X.Bytes()[12:32])

		fmt.Printf("New Private Key: %x\n", privKey.D.Bytes())
		fmt.Printf("Wallet Address:  %x\n", address[:])
		fmt.Printf("Store these securely.\n")
		os.Exit(0)
	}

	// Handle Transaction / Wallet Client Mode
	if *txMode {
		if *privKeyHex == "" || *toAddr == "" {
			log.Fatal("Missing required flags: --privkey and --to")
		}

		// 1. Load the Private Key
		privKey, err := crypto.HexToPrivateKey(*privKeyHex)
		if err != nil {
			log.Fatalf("Invalid private key: %v", err)
		}

		// Calculate sender address from public key (simplified 20-byte extraction)
		// In a real build, you hash the pubkey. For now, we take the last 20 bytes of X.
		var senderAddr types.Address
		copy(senderAddr[:], privKey.PublicKey.X.Bytes()[12:32])

		// 2. Parse Amount
		amount := new(big.Int)
		amount.SetString(*amountStr, 10)

		// 3. Construct the Transaction
		tx := types.Transaction{
			Sender:    senderAddr,
			Recipient: types.HexToAddress(*toAddr),
			Amount:    amount,
			Nonce:     *nonce,
			Data:      []byte(*dataStr),
		}

		// 4. Sign the Transaction
		payload := tx.Payload() // You defined this in Phase 2
		signature, err := crypto.SignData(privKey, payload)
		if err != nil {
			log.Fatalf("Failed to sign transaction: %v", err)
		}
		tx.Signature = signature

		// Generate a simple hash for the transaction (for the mempool ID)
		hash := sha256.Sum256(append(payload, signature...))
		tx.Hash = hash[:]

		// 5. Send via HTTP POST
		txJSON, _ := json.Marshal(tx)
		resp, err := http.Post(fmt.Sprintf("http://localhost%s/tx", *apiAddr), "application/json", bytes.NewBuffer(txJSON))
		if err != nil {
			log.Fatalf("Failed to connect to node API: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusAccepted {
			fmt.Printf("Transaction successfully submitted to mempool!\nHash: %x\n", tx.Hash[:4])
		} else {
			fmt.Printf("Node rejected transaction. HTTP Status: %d\n", resp.StatusCode)
		}
		os.Exit(0)
	}

	// Handle Node Startup
	if *nodeMode {
		log.Println("Starting Enterprise Blockchain Node...")

		// 1. Initialize Database (Phase 2)
		db, err := database.NewStore(*dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()

		// 2. Setup Proof-of-Authority (Phase 1)
		// Hardcode the genesis validator for testing
		genesisValidator := types.HexToAddress("0000000000000000000000000000000000000000")
		poaEngine := consensus.NewPoAEngine([]types.Address{genesisValidator})

		// 3. Initialize Blockchain Engine (Phase 4)
		blockchain := core.NewBlockchain(db, poaEngine)

		// --- NEW: GENESIS ALLOCATION ---
		// We use a specific address for you. (In reality, use the public address from --genkey)
		myAddress := types.HexToAddress("8bce22ec065205b4640de643b28496e5a613e099")
		// Check if we already have a balance to ensure we don't re-fund on every restart
		balance, _ := blockchain.GetBalance(myAddress)
		if balance.Cmp(big.NewInt(0)) == 0 {
			log.Println("Initializing Genesis State...")

			// Give ourselves 1,000,000 tokens
			genesisAccount, _ := blockchain.StateDB().GetAccount(myAddress)
			genesisAccount.AddBalance(big.NewInt(1000000))
			blockchain.StateDB().SaveAccount(genesisAccount)

			log.Printf("Allocated 1,000,000 tokens to Genesis Address: %x\n", myAddress[:4])
		}

		// 4. Initialize Mempool (Phase 3)
		txPool := mempool.NewMempool()

		// 5. Start P2P Network Server (Phase 3)
		p2pServer := network.NewServer(*p2pAddr, txPool, blockchain)
		go func() {
			if err := p2pServer.Start(); err != nil {
				log.Fatalf("P2P Server failed: %v", err)
			}
		}()

		// --- NEW: CONNECT TO PEER ---
		if *peerAddr != "" {
			go func() {
				// Give our own server a second to boot up first
				// In a production app, use proper synchronization instead of sleep
				importTime := true // Just a mental note: ensure "time" is imported!
				_ = importTime

				// Quick sleep to let the listener bind
				// (Make sure "time" is in your imports at the top of main.go)
				fmt.Println("Dialing peer...")
				if err := p2pServer.ConnectToPeer(*peerAddr); err != nil {
					log.Printf("Failed to connect to peer: %v", err)
				}
			}()
		}
		// ----------------------------

		// 6. Start REST API (Phase 5)
		apiServer := api.NewServer(*apiAddr, blockchain, txPool)
		go func() {
			if err := apiServer.Start(); err != nil {
				log.Fatalf("API Server failed: %v", err)
			}
		}()

		// --- NEW: START MINTER ---
		// We pass the genesis validator address we authorized in Phase 1
		go startMinter(blockchain, txPool, p2pServer, genesisValidator)
		// -------------------------

		// 7. Graceful Shutdown Listener
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Shutting down node gracefully...")
		os.Exit(0)

	}

	fmt.Println("Please specify a command. Use --help for options.")
}

// startMinter acts as the PoA block generator.
func startMinter(bc *core.Blockchain, pool *mempool.Mempool, p2p *network.Server, validatorAddr types.Address) {
	// Mint a new block every 10 seconds
	ticker := time.NewTicker(10 * time.Second)

	for {
		<-ticker.C

		// 1. Get pending transactions
		txs := pool.GetPending()

		// 2. We only mint a block if there are transactions
		if len(txs) == 0 {
			continue
		}

		log.Printf("Minting new block with %d transactions...\n", len(txs))

		// 3. Construct the Block Header
		header := types.BlockHeader{
			Version:   1,
			Timestamp: uint64(time.Now().Unix()),
			Number:    bc.GetTipHeight() + 1, // Fixed height logic
			TxRoot:    crypto.ComputeMerkleRoot(txs),
			Signer:    validatorAddr,
		}

		// 5. Assemble the complete block
		block := types.Block{
			Header:       header,
			Transactions: txs,
		}

		// 6. Process it locally
		if err := bc.ProcessBlock(block); err != nil {
			log.Printf("Failed to mint block: %v\n", err)
			continue
		}

		// 7. Clear the mempool of the transactions we just included
		pool.Clear(txs)

		// 8. Gossip the new block to the network
		var buf bytes.Buffer
		gob.NewEncoder(&buf).Encode(network.BlockMessage{Block: block})

		p2p.Broadcast(network.Envelope{
			Type: network.MsgBlock,
			Data: buf.Bytes(),
		})
	}
}
