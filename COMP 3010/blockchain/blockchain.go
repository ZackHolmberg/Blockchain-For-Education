package blockchain

type ConsensusComponent interface {
	consensusMethod()
}

type ProofComponent interface {
	proofMethod()
	validateProof()
}

type CommunicationComponent interface {
	getPeerChains()
	pingNetwork()
	recieveFromClient()
	sendToClient()
	recieveFromNetwork()
	broadcastToNetwork()
}

type Data interface {
	getData()
	toString()
}

type Blockchain struct {

	communicationComponent CommunicationComponent;
	proofComponent ProofComponent;
	consensusComponent ConsensusComponent;
	block Block;

}

type Block struct {
    
}

type LongestChain struct {

}

type ProofOfWork struct {
   
}

type Communicator struct {
  
}

type Transaction struct {
 
}