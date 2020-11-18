package main

import "fmt"

// ============================ Blockchain ============================

type Blockchain struct {

	communicationComponent CommunicationComponent
	proofComponent ProofComponent
	consensusComponent ConsensusComponent
	genesisBlock *Block

}

// ============================ Block ============================

type Block struct {

	index int
	timestamp string
	data Data
	prevHash string
	hash string
    
}

func (b Block) mine() {
    
}

type Data interface {
	getData()
	toString()
}

type Transaction struct {
	from string
	to string
	amount int
}

func (t Transaction) getData() {
}

func (t Transaction) toString()  {
}

// ============================ Consensus ============================

type ConsensusComponent interface {
	consensusMethod()
}

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

type ProofComponent interface {
	proofMethod()
	validateProof()
}

type ProofOfWork struct {
	nonce int
}

func (p ProofOfWork) proofMethod() {
    proofOfWork()
}

func (p ProofOfWork) validateProof() {
}

func proofOfWork() int{
    return 0
}

// ============================ Communication ============================

type CommunicationComponent interface {
	getPeerChains()
	pingNetwork()
	recieveFromClient()
	sendToClient()
	recieveFromNetwork()
	broadcastToNetwork()
}

type Communicator struct {
  
}

func (c Communicator) getPeerChains()  {
}

func (c Communicator) pingNetwork()  {
}

func (c Communicator) recieveFromClient()  {
}

func (c Communicator) sendToClient()  {
}

func (c Communicator) recieveFromNetwork()  {
}

func (c Communicator) broadcastToNetwork() {
}

// ============================ Main ============================

func main() {

	communicator := Communicator{}
	proofOfWork := ProofOfWork{}
	longestChain := LongestChain{}

	blockchain := Blockchain{communicator, proofOfWork, longestChain, &Block{0,"0",Transaction{"zack","zack2",42},"0","0"}}

	fmt.Printf("\n%#v\n\n",blockchain)
}
