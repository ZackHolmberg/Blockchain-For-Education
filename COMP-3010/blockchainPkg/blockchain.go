package blockchain

// ============================ Blockchain ============================

// Blockchain is the Blockchain object
type Blockchain struct {
	communicationComponent CommunicationComponent
	proofComponent         ProofComponent
	consensusComponent     ConsensusComponent
	genesisBlock           *Block
}

// ============================ Block ============================

// Block is the Block object
type Block struct {
	index     int
	timestamp string
	data      Data
	prevHash  string
	hash      string
}

func (b Block) mine() {

}

// Data is an interface used to standardize methods for any type of Block data
type Data interface {
	getData()
	toString()
}

// Transaction is a type of Block data
type Transaction struct {
	from   string
	to     string
	amount int
}

func (t Transaction) getData() {
}

func (t Transaction) toString() {
}

// ============================ Consensus ============================

// ConsensusComponent standardizes methods for any Blockchain consensus component
type ConsensusComponent interface {
	consensusMethod()
}

// LongestChain algorithm used in Blockchain consensus
type LongestChain struct {
	peerChains []Blockchain
}

func (l LongestChain) consensusMethod() {
	longestChain()
}

func longestChain() int {
	return 0
}

// ============================ Proof ============================

//ProofComponent standardizes methods for any Blockchain proof component
type ProofComponent interface {
	proofMethod()
	validateProof()
}

// ProofOfWork algorithm used in mining blocks
type ProofOfWork struct {
	nonce int
}

func (p ProofOfWork) proofMethod() {
	proofOfWork()
}

func (p ProofOfWork) validateProof() {
}

func proofOfWork() int {
	return 0
}

// ============================ Communication ============================

// CommunicationComponent standardizes methods for any Blockchain communcation component
type CommunicationComponent interface {
	getPeerChains()
	pingNetwork()
	recieveFromClient()
	sendToClient()
	recieveFromNetwork()
	broadcastToNetwork()
}

// Communicator implements CommunicationsComponent and facilities Blockchain communication
type Communicator struct {
}

func (c Communicator) getPeerChains() {
}

func (c Communicator) pingNetwork() {
}

func (c Communicator) recieveFromClient() {
}

func (c Communicator) sendToClient() {
}

func (c Communicator) recieveFromNetwork() {
}

func (c Communicator) broadcastToNetwork() {
}
