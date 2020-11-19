package blockchain

// // ============================ Blockchain ============================

// // Blockchain is the Blockchain object
// type Blockchain struct {
// 	CommunicationComponent CommunicationComponent
// 	ProofComponent         ProofComponent
// 	ConsensusComponent     ConsensusComponent
// 	GenesisBlock           *Block
// }

// // CreateGenesisBlock add the genesis block to the blockchain
// func (b Blockchain) CreateGenesisBlock() {
// 	block := &Block{}
// 	b.GenesisBlock = block
// }

// // ============================ Block ============================

// // Block is the Block object
// type Block struct {
// 	Index     int
// 	Timestamp string
// 	Data      Data
// 	PrevHash  string
// 	Hash      string
// }

// func (b Block) Mine() {

// }

// // Data is an interface used to standardize methods for any type of Block data
// type Data interface {
// 	GetData()
// 	ToString()
// }

// // Transaction is a type of Block data
// type Transaction struct {
// 	From   string
// 	To     string
// 	Amount int
// }

// func (t Transaction) GetData() {
// }

// func (t Transaction) ToString() {
// }

// // ============================ Consensus ============================

// // ConsensusComponent standardizes methods for any Blockchain consensus component
// type ConsensusComponent interface {
// 	ConsensusMethod()
// }

// // LongestChain algorithm used in Blockchain consensus
// type LongestChain struct {
// 	PeerChains []Blockchain
// }

// func (l LongestChain) ConsensusMethod() {
// 	longestChain()
// }

// func longestChain() int {
// 	return 0
// }

// // ============================ Proof ============================

// //ProofComponent standardizes methods for any Blockchain proof component
// type ProofComponent interface {
// 	ProofMethod()
// 	ValidateProof()
// }

// // ProofOfWork algorithm used in mining blocks
// type ProofOfWork struct {
// 	nonce int
// }

// func (p ProofOfWork) ProofMethod() {
// 	proofOfWork()
// }

// func (p ProofOfWork) ValidateProof() {
// }

// func proofOfWork() int {
// 	return 0
// }

// // ============================ Communication ============================

// // CommunicationComponent standardizes methods for any Blockchain communcation component
// type CommunicationComponent interface {
// 	GetPeerChains()
// 	PingNetwork()
// 	RecieveFromClient()
// 	SendToClient()
// 	RecieveFromNetwork()
// 	BroadcastToNetwork()
// }

// // Communicator implements CommunicationsComponent and facilities Blockchain communication
// type Communicator struct {
// }

// func (c Communicator) GetPeerChains() {
// }

// func (c Communicator) PingNetwork() {
// }

// func (c Communicator) RecieveFromClient() {
// }

// func (c Communicator) SendToClient() {
// }

// func (c Communicator) RecieveFromNetwork() {
// }

// func (c Communicator) BroadcastToNetwork() {
// }
