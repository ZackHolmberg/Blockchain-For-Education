package main

import (
	"blockchain"
	"fmt"
)

var communicator *blockchain.Communicator
var proofOfWork blockchain.ProofOfWork
var longestChain blockchain.LongestChain

func init() {

	communicator = &blockchain.Communicator{}
	proofOfWork = blockchain.ProofOfWork{ProofDifficulty: 2}
	longestChain = blockchain.LongestChain{}
}

// ============================ Main ============================

func main() {

	// start blockchain
	fmt.Println("\nStarting Blockchain...")
	bc, err := blockchain.NewBlockchain(communicator, proofOfWork, longestChain)
	if err != nil {
		fmt.Printf("Fatal error creating Blockchain: %#v\n", err)
	} else {
		bc.Run()
	}

}
