package main

import (
	"blockchain"
	"fmt"
)

// ============================ Main ============================

func main() {

	communicator := blockchain.Communicator{}
	proofOfWork := blockchain.ProofOfWork{}
	longestChain := blockchain.LongestChain{}
	blockchain := blockchain.NewBlockchain(communicator, proofOfWork, longestChain)

	fmt.Printf("\n%#v\n\n", blockchain)
}

// --- Misc. Notes ---
// Add a makefile
// Middleware/MessageHandler Proxy on the network that students send HTTPS requests to, which it then transalates
// to seomething readable by the peers on the network and broadcasts
