# Name: Zack Holmberg  
# Student #: 7777823  
# UserID: holmbezt
# Credit to this article: https://medium.com/crypto-currently/lets-make-the-tiniest-blockchain-bigger-ac360a328f4d
# which provided direction and some code snippets on how to go about implementing a very basic cryptocurrency

import json
import datetime as date
import requests
import hashlib as hasher
from flask import Flask
from flask import request
node = Flask(__name__)

# Define a ZackCoin block
class ZackCoin:
    def __init__(self, index, timestamp, data, prevHash):
        self.index = index
        self.timestamp = timestamp
        self.data = data
        self.prevHash = prevHash
        self.hash = self.hashFunction()

    def hashFunction(self):
        sha = hasher.sha256()
        sha.update(str(self.index) + str(self.timestamp) +
                   str(self.data) + str(self.prevHash))
        return sha.hexdigest()

# Create the Genesis ZackCoin (different case than adding a ZackCoin to the chain)
def createGenesisNode():
    # Create a ZackCoin with index 0 and previous hash of 0, since there are no previous hashes
    return ZackCoin(0, date.datetime.now(), {
        "proofOfWork": 42,
        "transactions": None
    }, "0")

# Fake address of a miner so that we can send them a reward for when they mine a ZackCoin
minerAddress = "randomMinerAddress"

# This node's copy of the chain
theZackCoinchain = []

# Create the genesis ZackCoin
theZackCoinchain.append(createGenesisNode())

# Track this node's transactions
transactions = []

# Track data of peer nodes on the network so that we can communicate with them
peerNodes = []

# The data of a ZackCoin will be a transaction record, similar to Bitcoin
@node.route('/transaction', methods=['POST'])
def transaction():

    # Get JSON from request
    newTransaction = request.get_json()

    # Add transaction to the list
    transactions.append(newTransaction)

    print("Received a new transaction on the network")
    print("From: {}".format(newTransaction['from'].encode('ascii', 'replace')))
    print("To: {}".format(newTransaction['to'].encode('ascii', 'replace')))
    print("Amount: {}\n".format(newTransaction['amount']))
    # Then we let the client know it worked out
    return "Successfully added new transaction.\n"


@node.route('/AllZackCoins', methods=['GET'])
def getAllZackCoins():
    theChain = theZackCoinchain

    for i in range(len(theChain)):
        ZackCoin = theChain[i]
        ZackCoinIndex = str(ZackCoin.index)
        ZackCoinTimestamp = str(ZackCoin.timestamp)
        ZackCoinData = str(ZackCoin.data)
        ZackCoinHash = ZackCoin.hash
        theChain[i] = {
            "index": ZackCoinIndex,
            "timestamp": ZackCoinTimestamp,
            "data": ZackCoinData,
            "hash": ZackCoinHash
        }
    theChain = json.dumps(theChain)
    return theChain


def getPeerChains():
    # Get the chain copies held by all other nodes on the network
    peerChains = []
    for nodeURL in peerNodes:
        # Get peer chains with a GET request
        peerChain = requests.get(nodeURL + "/ZackCoins").content
        # Convert the JSON object to a Python dictionary
        peerChain = json.loads(peerChain)
        # Add it to local chain
        peerChains.append(peerChain)
    return peerChains


# Uses the longest chain rule as a rule for consensus amongst Nodes on the network
def consensusAlgorithm():
    # Get the chain copies from other nodes
    peerChains = getPeerChains()

    # Apply the longest chain rule to get longest check on the network
    longestChain = theZackCoinchain
    for peerChain in peerChains:
        if len(longestChain) < len(peerChain):
            longestChain = peerChain

    theZackCoinchain = longestChain


def proofOfWork(lastProof):

    i = lastProof + 1
    # Find a number to satisfy the proof of work. For ZackCoin,
    # a suitable solution is any number divisble by 42 (obviously)
    while not (i % 42 == 0 and i % lastProof == 0):
        i += 1
    # Once a suitable number to satisfy the proof of work is found, return it
    return i


@node.route('/mine', methods=['GET'])
def mine():

    # Get previous proof of work
    lastZackCoin = theZackCoinchain[len(theZackCoinchain) - 1]
    lastProof = lastZackCoin.data['proofOfWork']

    # Find a proof of work for the ZackCoin being mined
    proof = proofOfWork(lastProof)

    # Once we find a valid proof of work, we know we can mine a ZackCoin so we reward the miner by sending them, and thus adding, a (sizeable) reward/transaction
    transactions.append(
        {"from": "network", "to": minerAddress, "amount": 42}
    )
    # Now we can gather the data needed to create the new ZackCoin
    newZackCoinData = {
        "proofOfWork": proof,
        "transactions": list(transactions)
    }
    newZackCoinIndex = lastZackCoin.index + 1
    newZackCoinTimestamp = date.datetime.now()
    lastZackCoinHash = lastZackCoin.hash

    # Clear the transaction list
    transactions[:] = []

    # Create new ZackCoin using data collected earlier
    newZackCoin = ZackCoin(
        newZackCoinIndex,
        newZackCoinTimestamp,
        newZackCoinData,
        lastZackCoinHash
    )
    # And finally add the new ZackCoin to the chain
    theZackCoinchain.append(newZackCoin)

    # Inform the client that a ZackCoin was mined successfully
    toReturn = json.dumps({
        "index": newZackCoinIndex,
        "timestamp": str(newZackCoinTimestamp),
        "data": newZackCoinData,
        "hash": lastZackCoinHash
    }) + "\n"

    return toReturn

# Run the Flask application
node.run()
