package main

import (
	"blockchain"
	"fmt"
)

var communicator *blockchain.Communicator
var proofOfWork blockchain.ProofOfWork

func init() {

	communicator = &blockchain.Communicator{}
	proofOfWork = blockchain.ProofOfWork{ProofDifficulty: 6}
}

// ============================ Main ============================

func main() {

	// start blockchain
	fmt.Println("\nStarting Blockchain Peer...")
	bc, err := blockchain.NewPeer(communicator, proofOfWork)
	if err != nil {
		fmt.Printf("Fatal error creating Blockchain Peer: %+v\n", err)
	} else {
		bc.Run()
	}

}
