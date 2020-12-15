package blockchain

import (
	"log"
)

// ============================ Consensus ============================

// LongestChain algorithm used in Blockchain consensus
type LongestChain struct {
	PeerChains []Blockchain
}

// Initialize is the interface method that calls this component's initialize method
func (l LongestChain) Initialize() error {
	// No initialization needed for this implementation
	return nil
}

// Terminate is the interface method that calls this component's cleanup method
func (l LongestChain) Terminate() {
	// No initialization needed for this implementation
}

// ConsensusMethod is the interface method that calls this component's consensus method, longestChain
func (l LongestChain) ConsensusMethod(p [][]Block, c []Block) ([]Block, error) {

	longestChain, err := longestChain(p, c)

	if err != nil {
		log.Printf("Error running conesnsus method: %+v\n", err)
		return nil, err
	}

	return longestChain, nil
}

// longestChain uses the longestChain algorithm to achieve blockchain consensus amongst network
func longestChain(p [][]Block, c []Block) ([]Block, error) {

	index := -1
	length := len(c)

	for i, chain := range p {
		currLength := len(chain)
		if currLength > length {
			index = i
			length = currLength
		}
	}

	var toReturn []Block

	if index != -1 {
		toReturn = p[index]
	} else {
		toReturn = c
	}

	log.Println("Ran consensus")

	return toReturn, nil
}
