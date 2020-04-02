package core

import (
	"sort"
	"time"
)

// A valid hexadecimal string that represents a user's public key. It must be on the ECDSA SECP256k1 curve.
type UserPublicKey string

// The amount of unspent coin each user has associated with their public key
type UTXO map[UserPublicKey]int

// A Blockchain is a struct that stores a Chain of Blocks, as well as MemPool and manages its own UTXO map.
// It also stores a ValidationServerURL and an Operator Public key which is used to identify that node when mining
type Blockchain struct {
	Chain   []Block       // The actual chain of transactions that makes up this "Blockchain"
	MemPool []Transaction // The waiting room of transactions that are yet to be incorporated in a block. These get cleared out every 24 hours.
	UTXO    UTXO          // The amount of unspent transactions each user has associated with their public key

	ValidationServerURL string        // A link to a server that can be used to validate signatures
	OperatorPublicKey   UserPublicKey // A public key that is used to identify the node when mining (so this node can receive mining rewards

	IsMining bool // Stores whether the node is mining or not. If the node is mining and this bool is set to false, the node will terminate its mining process.
}

// Adds a transaction to the MemPool (but will do nothing to incorporate it into a block or verify it).
func (b *Blockchain) AddTransactionToMemPool(transaction Transaction) {
	b.MemPool = append(b.MemPool, transaction)
}

// Adds a new block to the chain (by first verifying it and getting its UTXO). It has side effects:
//  - It stops all mining processes on this node
//	- It removes the transactions inside the block from the MemPool
//  - It updates the UTXO
func (b *Blockchain) AddMinedBlockToChain(block Block) bool {
	// Cancel mining processes as a new block has been found
	b.IsMining = false

	tempChain := append(b.Chain, block)

	isValid, utxo := ValidateChain(tempChain, b.ValidationServerURL)

	if isValid {
		// Clear Mempool of confirmed transactions (transactions that are now in this block)
		b.MemPool = RemoveConfirmedTransactions(b.MemPool, block.Transactions)

		// Update UTXO
		b.UTXO = utxo

		// Update chain
		b.Chain = tempChain

		return true
	} else {
		return false
	}
}

// Finds a valid proof for a block and validates transactions from the MemPool. It removes invalid transactions.
// It returns a pointer to a new block that will be nil if the mining process was canceled.
// It does not add this block to the chain itself.
func (b Blockchain) MineBlock(shouldMine *bool) *Block {
	// Ensure that we are mining
	*shouldMine = true

	// Add a "coinbase" transaction that mints 10 coins to the miner (this node's public key)
	newTransactions := append(b.MemPool, Transaction{"0", b.OperatorPublicKey, 0, 10, ""})

	// Sort the transactions by their timestamp
	sort.Slice(newTransactions, func(index1, index2 int) bool {
		return newTransactions[index1].Timestamp < newTransactions[index2].Timestamp
	})

	// Remove invalid transactions
	for index, transaction := range newTransactions {
		// If the transaction is a coinbase transaction (the first transaction) or the sender has enough coin then update their UTXO
		if (transaction.Sender == "0" && index == 0) || (transaction.Amount < b.UTXO[transaction.Sender] && ValidateSignature(transaction, b.ValidationServerURL)) {
			b.UTXO[transaction.Sender] -= transaction.Amount
			b.UTXO[transaction.Recipient] += transaction.Amount
		} else {
			newTransactions = RemoveFromTransactions(newTransactions, index)
		}

	}

	blockHeader := BlockHeader{time.Now().Unix(), newTransactions, LastBlock(b.Chain).hash()}

	// Create a proof with the appropriate difficulty
	proof := Proof{Nonce: 0, DifficultyThreshold: DetermineDifficultyForChainIndex(b.Chain, len(b.Chain))}

	// Keep incrementing the nonce until we have a valid proof
	for !ValidateProof(Block{blockHeader, proof}) {
		// Cancel mining if we are having this mine terminated
		if *shouldMine == false {
			return nil
		}
		proof.Nonce += 1
	}

	// We found a valid block!
	return &Block{blockHeader, proof}
}

// Does these checks to ensure the chain is valid:
//  - Check that previous hashes are valid
//  - Check that users have enough UTXO to afford transactions
//  - Check that proofs are valid
//  - Check that there are not more than one coinbase transaction in each block
//  - Check that signatures are valid
//  - Check that difficulty threshold is valid
// It returns whether the chain is valid and an updated UTXO (or nil if not valid).
func ValidateChain(blocks []Block, validationServerURL string) (bool, UTXO) {
	utxo := make(UTXO)

	// Starting at the 2nd block, iterate over all blocks (so we don't verify the genesis block, only that it hasn't changed)
	for index := 1; index < len(blocks); index++ {
		block := blocks[index]
		lastBlock := blocks[index-1]

		// Check previous hash is valid and that proof is valid
		if block.PreviousHash != lastBlock.hash() || !ValidateProof(block) {
			return false, nil
		}

		// Check that difficulty threshold is valid
		if block.Proof.DifficultyThreshold != DetermineDifficultyForChainIndex(blocks, index) {
			return false, nil
		}

		// Check the transactions in it are valid
		for index, transaction := range block.Transactions {
			// If the transaction is a coinbase transaction (the first transaction) or the sender has enough coin: update their UTXO
			if (transaction.Sender == "0" && index == 0) || (transaction.Amount < utxo[transaction.Sender] && ValidateSignature(transaction, validationServerURL)) {
				utxo[transaction.Sender] -= transaction.Amount
				utxo[transaction.Recipient] += transaction.Amount
			} else {
				return false, nil
			}
		}
	}

	return true, utxo
}

// A Block is a block header with a proof that when put into the format {Proof}-{BlockHeader}, can be hashed into a hex string with x leading 0s.
type Block struct {
	BlockHeader
	Proof Proof // The nonce and difficulty threshold that validates this block
}

// A BlockHeader stores a timestamp, a list of transactions and the hash of the previous block.
type BlockHeader struct {
	Timestamp    int64         // The time when this block header was generated
	Transactions []Transaction // The transactions the enclosing block validates
	PreviousHash string        // The hash of the previous block
}

// A transaction stores information about a transaction with a signature.
type Transaction struct {
	Sender    UserPublicKey // The public key of the sender
	Recipient UserPublicKey // The public key of the recipient
	Amount    int           // The amount of coin transferred
	Timestamp int64         // The time at which this transaction was made. This value does not need to be accurate, it is only for the purpose of ordering transactions in a BlockHeader.
	Signature string        // A hex string that is an ECDSA signed representation of this transaction ({SENDER} -{AMOUNT}-> {RECIPIENT})
}
