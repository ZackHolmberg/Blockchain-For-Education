package main

import (
	"fmt"

	"github.com/ZackHolmberg/Blockchain-Honours-Project/tree/main/COMP-3010/blockchain"
)

// ============================ Main ============================

func main() {

	// communicator := Communicator{}
	// proofOfWork := ProofOfWork{}
	// longestChain := LongestChain{}

	// blockchain := Blockchain{communicator, proofOfWork, longestChain, &Block{0, "0", Transaction{"zack", "zack2", 42}, "0", "0"}}
	// blockchain := "hello, world!"
	blockchain := blockchain.Blockchain{}
	fmt.Printf("\n%#v\n\n", blockchain)
}
