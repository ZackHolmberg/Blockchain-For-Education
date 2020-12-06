package main

import (
	"blockchain"
	"fmt"
	"sync"
)

var communicator *blockchain.Communicator

func init() {

	communicator = &blockchain.Communicator{}
}

// ============================ Main ============================

func main() {

	wg := new(sync.WaitGroup)

	// start middleware
	fmt.Println("\nStarting Middleware...")
	m := blockchain.NewMiddleware(communicator)
	wg.Add(1)
	go m.Run(wg)

	wg.Wait()

}

// --- Misc. Notes ---
// Middleware/MessageHandler Proxy on the network that students send HTTPS requests to, which it then transalates
// to something readable by the peers on the network and broadcasts
// students might have to go get zeroconf pkg, or maybe we can install it on aviary
