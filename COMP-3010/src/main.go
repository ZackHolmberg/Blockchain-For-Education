package main

import (
	"blockchain"
	"fmt"
	"sync"
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

	wg := new(sync.WaitGroup)

	// start middleware
	fmt.Println("\nStarting Middleware...")
	m := blockchain.NewMiddleware(communicator)
	wg.Add(1)
	go m.Run(wg)
	// time.Sleep(5 * time.Second)

	// // start blockchain
	// fmt.Println("\nStarting Blockchain...")
	// bc := blockchain.NewBlockchain(communicator, proofOfWork, longestChain)
	// // fmt.Printf("\nThe Blockchain: \n%#v\n\n", bc)
	// wg.Add(1)
	// go bc.Run(wg)

	wg.Wait()

	// hash := proofOfWork.ProofMethod(*bc.GenesisBlock)
	// fmt.Printf("\nHash: %#v\n\n", hash)

	// transaction := blockchain.Transaction{From: "John", To: "Doe", Amount: 42}
	// bc.Mine(transaction)
	// fmt.Printf("\nThe Blockchain after mining: \n%#v\n\n", bc)

}

// --- Misc. Notes ---
// Middleware/MessageHandler Proxy on the network that students send HTTPS requests to, which it then transalates
// to something readable by the peers on the network and broadcasts
// students might have to go get zeroconf pkg, or maybe we can install it on aviary
