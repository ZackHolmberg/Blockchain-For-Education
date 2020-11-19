package blockchain

// ============================ Proof ============================

//ProofComponent standardizes methods for any Blockchain proof component
type ProofComponent interface {
	ProofMethod()
	ValidateProof()
}

// ProofOfWork algorithm used in mining blocks
type ProofOfWork struct {
	nonce int
}

func (p ProofOfWork) ProofMethod() {
	proofOfWork()
}

func (p ProofOfWork) ValidateProof() {
}

func proofOfWork() int {
	return 0
}
