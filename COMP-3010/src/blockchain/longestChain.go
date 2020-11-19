package blockchain

// ============================ Consensus ============================

// LongestChain algorithm used in Blockchain consensus
type LongestChain struct {
	PeerChains []Blockchain
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
