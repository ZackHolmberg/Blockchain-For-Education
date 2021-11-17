package main

import (
	"fmt"

	"github.com/ZackHolmberg/Blockchain-Honours-Project/COMP-3010/src/blockchain"
)

var communicator *blockchain.Communicator
var proofOfWork blockchain.ProofOfWork
var longestChain blockchain.LongestChain

func init() {

	communicator = &blockchain.Communicator{}
	proofOfWork = blockchain.ProofOfWork{ProofDifficulty: 5}
	longestChain = blockchain.LongestChain{}
}

// ============================ Main ============================

func main() {

	// start blockchain
	fmt.Println("\nStarting Blockchain...")
	bc, err := blockchain.NewBlockchain(communicator, proofOfWork, longestChain)
	if err != nil {
		fmt.Printf("Fatal error creating Blockchain: %+v\n", err)
	} else {
		bc.Run()
	}

}
