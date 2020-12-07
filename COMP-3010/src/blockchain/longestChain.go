package blockchain

import "log"

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
func (l LongestChain) ConsensusMethod() error {

	err := longestChain()

	if err != nil {
		log.Printf("Error running conesnsus method: %#v\n", err)
		return err
	}

	return nil
}

// longestChain uses the longestChain algorithm to achieve blockchain consensus amongst network
func longestChain() error {
	//TODO: Implement
	return nil
}
