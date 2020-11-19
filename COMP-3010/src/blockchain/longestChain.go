package blockchain

// ============================ Consensus ============================

// ConsensusComponent standardizes methods for any Blockchain consensus component
type ConsensusComponent interface {
	ConsensusMethod()
}

// LongestChain algorithm used in Blockchain consensus
type LongestChain struct {
	PeerChains []Blockchain
}

func (l LongestChain) ConsensusMethod() {
	longestChain()
}

func longestChain() int {
	return 0
}
