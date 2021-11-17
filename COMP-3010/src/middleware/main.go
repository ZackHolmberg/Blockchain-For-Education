package main

import (
	"fmt"

	"github.com/ZackHolmberg/Blockchain-Honours-Project/COMP-3010/src/blockchain"
)

var communicator *blockchain.Communicator

func init() {

	communicator = &blockchain.Communicator{}
}

// ============================ Main ============================

func main() {

	// start middleware
	fmt.Println("\nStarting Middleware...")
	m, err := blockchain.NewMiddleware(communicator, 8080, 8090)
	if err != nil {
		fmt.Printf("Fatal error creating Middleware: %+v\n", err)
	} else {
		m.Run()
	}

}
