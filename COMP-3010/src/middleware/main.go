package main

import (
	"blockchain"
	"fmt"
)

var communicator *blockchain.Communicator

func init() {

	communicator = &blockchain.Communicator{}
}

// ============================ Main ============================

func main() {

	// start middleware
	fmt.Println("\nStarting Blockchain Middleware...")
	m, err := blockchain.NewMiddleware(communicator, 8080, 8090)
	if err != nil {
		fmt.Printf("Fatal error creating Blockchain Middleware: %+v\n", err)
	} else {
		m.Run()
	}

}
