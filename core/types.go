package core

import (
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/kademlia"
)

// The reward given to miners for mining a block
var coinbaseReward int64 = 1000

// The first block in our Blockchain
var GenesisBlock = Block{BlockHeader: BlockHeader{Timestamp: 1585852979, Transactions: []Transaction{Transaction{Sender: "0", Recipient: "b61e63485c4782d6495aa0091c6785d8b6c0a945a23d9b158093bbf3d93d6bb9024e6cab467cc11b51e1b1a158637a778473418298b09a7dd39c148863b1833c", Amount: 100000000000000, Timestamp: 1585852961, Signature: ""}}, PreviousHash: ""}, Proof: Proof{Nonce: 0, DifficultyThreshold: 0}}

// The amount of unspent coin each user has associated with their public key
type UTXO map[string]int64

// A Blockchain is a struct that stores a Chain of Blocks, as well as MemPool and manages its own UTXO map.
// It also stores a ValidationServerURL and an Operator Public key which is used to identify that node when mining
type LocalNode struct {
	Chain   []Block       // The actual chain of transactions that makes up this "Blockchain"
	MemPool []Transaction // The waiting room of transactions that are yet to be incorporated in a block. These get cleared out every 24 hours.
	UTXO    UTXO          // The amount of unspent transactions each user has associated with their public key

	ValidationServerURL string // A link to a server that can be used to validate signatures
	OperatorPublicKey   string // A public key that is used to identify the node when mining (so this node can receive mining rewards

	IsMining bool // Stores whether the node is mining or not. If the node is mining and this bool is set to false, the node will terminate its mining process.

	node             *noise.Node        // This node's P2P representation
	kademliaProtocol *kademlia.Protocol // Stores this block's peers

	incomingChains            chan []Block // Stores incoming chains for our consensus algorithm
	MinimumChainsForConsensus int          // How many chains we need before we run consensus
}

// A Block is a block header with a proof that when put into the format {Proof}-{BlockHeader}, can be hashed into a hex string with x leading 0s.
type Block struct {
	BlockHeader
	Proof Proof // The nonce and difficulty threshold that validates this block
}

// The nonce and difficulty threshold achieved by the nonce and BlockHeader to generate proof of work.
type Proof struct {
	Nonce               int64 // The random factor that changes the hash
	DifficultyThreshold int64 // The number of leading 0s required in the hash
}

// A BlockHeader stores a timestamp, a list of transactions and the hash of the previous block.
type BlockHeader struct {
	Timestamp    int64         // The time when this block header was generated
	Transactions []Transaction // The transactions the enclosing block validates
	PreviousHash string        // The hash of the previous block
}

// A transaction stores information about a transaction with a signature.
type Transaction struct {
	Sender    string // The public key of the sender (ECDSA SECP256k1)
	Recipient string // The public key of the recipient (ECDSA SECP256k1)
	Amount    int64  // The amount of coin transferred
	Timestamp int64  // The time at which this transaction was made. This value does not need to be accurate, it is only for the purpose of ordering transactions in a BlockHeader.
	Signature string // A hex string that is an ECDSA signed representation of this transaction ({SENDER} -{AMOUNT}-> {RECIPIENT} ({TIMESTAMP}))
}
