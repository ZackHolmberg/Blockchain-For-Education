package main

import (
	"blockchain"
	"fmt"
	"log"
	"os"
	"os/exec"
)

var communicator blockchain.Communicator
var proofOfWork blockchain.ProofOfWork
var longestChain blockchain.LongestChain

func init() {
	serviceName, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}

	//>>> TODO: Each node needs a unique port or host. Can't be same host and port. How to assign an unused port dynamically?
	communicator = blockchain.NewCommunicator(fmt.Sprintf("%s-%s", serviceName, out), "_blockchain-P2P-Network._udp", "local.", 42424)
	proofOfWork = blockchain.ProofOfWork{ProofDifficulty: 4}
	longestChain = blockchain.LongestChain{}
}

// ============================ Main ============================

func main() {

	bc := blockchain.NewBlockchain(communicator, proofOfWork, longestChain)

	fmt.Printf("\nThe Blockchain: \n%#v\n\n", bc)
	hash := proofOfWork.ProofMethod(*bc.GenesisBlock)
	fmt.Printf("\nHash: %#v\n\n", hash)

	transaction := blockchain.Transaction{From: "John", To: "Doe", Amount: 42}
	bc.Mine(transaction)
	fmt.Printf("\nThe Blockchain after mining: \n%#v\n\n", bc)

	//Consider whether this should be apart of a cleanupBlockchain function
	for true {
	}
	communicator.TerminateCommunicator()

}

// --- Misc. Notes ---
// Middleware/MessageHandler Proxy on the network that students send HTTPS requests to, which it then transalates
// to seomething readable by the peers on the network and broadcasts
// students might have to go get zeroconf pkg, or maybe we can install it on aviary
