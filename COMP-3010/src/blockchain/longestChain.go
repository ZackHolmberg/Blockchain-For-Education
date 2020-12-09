package blockchain

import (
	"fmt"
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
func (l LongestChain) ConsensusMethod(c CommunicationComponent) ([]Block, error) {

	longestChain, err := longestChain(c)

	if err != nil {
		log.Printf("Error running conesnsus method: %+v\n", err)
		return nil, err
	}

	return longestChain, nil
}

// longestChain uses the longestChain algorithm to achieve blockchain consensus amongst network
func longestChain(c CommunicationComponent) ([]Block, error) {

	var index int
	length := 0

	peerChains, err := c.GetPeerChains()
	if err != nil {
		log.Printf("Error getting peer chains: %+v\n", err)
		return nil, err
	}

	for i := 0; i < len(peerChains); i++ {
		if len(peerChains[i]) > length {
			index = i
			length = len(peerChains[i])
		}
	}

	fmt.Println("<<Ran consensus>>")

	return peerChains[index], nil
}
