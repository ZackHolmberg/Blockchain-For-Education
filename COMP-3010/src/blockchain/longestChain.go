package blockchain

// ============================ Consensus ============================

// LongestChain algorithm used in Blockchain consensus
type LongestChain struct {
	PeerChains []Blockchain
}

// Initialize is the interface method that calls this component's initialize method
func (l LongestChain) Initialize() {
	// No initialization needed for this implementation
}

// Terminate is the interface method that calls this component's cleanup method
func (l LongestChain) Terminate() {
	// No initialization needed for this implementation
}

// ConsensusMethod is the interface method that calls this component's consensus method, longestChain
func (l LongestChain) ConsensusMethod() {
	longestChain()
}

// longestChain uses the longestChain algorithm to achieve blockchain consensus amongst network
func longestChain() int {
	//TODO: Implement
	return 0
}
