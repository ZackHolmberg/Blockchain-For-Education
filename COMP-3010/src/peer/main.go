package main

import (
	"blockchain"
	"fmt"
)

var communicator *blockchain.Communicator
var proofOfWork *blockchain.ProofOfWork
var proofOfStake *blockchain.ProofOfStake
var client *blockchain.Client

func init() {

	communicator = &blockchain.Communicator{}
	proofOfWork = &blockchain.ProofOfWork{ProofDifficulty: 6}
	proofOfStake = &blockchain.ProofOfStake{}
	client = &blockchain.Client{}
}

// ============================ Main ============================

func main() {

	// start blockchain
	fmt.Println("\nStarting Blockchain Peer...")
	// bc, err := blockchain.NewPeer(communicator, proofOfWork, client)
	bc, err := blockchain.NewPeer(communicator, proofOfStake, client)
	if err != nil {
		fmt.Printf("Fatal error creating Blockchain Peer: %+v\n", err)
	} else {
		bc.Run()
	}

}
