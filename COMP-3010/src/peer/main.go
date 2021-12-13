package main

import (
	"blockchain"
	"fmt"
)

var communicator *blockchain.Communicator
var proofOfWork *blockchain.ProofOfWork
var proofOfStake *blockchain.ProofOfStake

func init() {

	communicator = &blockchain.Communicator{}
	proofOfWork = &blockchain.ProofOfWork{ProofDifficulty: 6}
	proofOfStake = &blockchain.ProofOfStake{}
}

// ============================ Main ============================

func main() {

	// start blockchain
	fmt.Println("\nStarting Blockchain Peer...")
	// bc, err := blockchain.NewPeer(communicator, proofOfWork)
	bc, err := blockchain.NewPeer(communicator, proofOfStake)
	if err != nil {
		fmt.Printf("Fatal error creating Blockchain Peer: %+v\n", err)
	} else {
		bc.Run()
	}

}
