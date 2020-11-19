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
	var blockchain = blockchain.Blockchain{CommunicationComponent: communicator, ProofComponent: proofOfWork, ConsensusComponent: longestChain, GenesisBlock: nil, Blockchain: nil}
	blockchain.CreateGenesisBlock()

	fmt.Printf("\n%#v\n\n", blockchain)
}
