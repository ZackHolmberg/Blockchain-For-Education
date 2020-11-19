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
