package core

import (
	"fmt"
	"reflect"
	"sort"
	"time"
)

var coinbaseReward = 1000

// Adds a transaction to the MemPool (but will do nothing to incorporate it into a block or verify it).
func (l *LocalNode) AddTransactionToMemPool(transaction Transaction) {
	// Run this in a separate Goroutine as IsTransactionAlreadyInMemPoolOrChain could take a while.
	go func() {
		// If transaction is not already in MemPool/Chain:
		if !IsTransactionAlreadyInMemPoolOrChain(transaction, l.MemPool, l.Chain) {
			// Add transaction to MemPool.
			l.MemPool = append(l.MemPool, transaction)
		}
	}()

}

// Adds a new block to the chain (by first verifying it and getting its UTXO). It has side effects:
//  - It stops all mining processes on this node
//	- It removes the transactions inside the block from the MemPool
//  - It updates the UTXO
func (l *LocalNode) AddMinedBlockToChain(block Block) bool {
	// Cancel mining processes as a new block has been found
	l.IsMining = false

	// Create a copy of the chain with the new block
	tempChain := append(l.Chain, block)

	// Check if that block is valid
	isValid, newUTXO := ValidateBlock(len(tempChain)-1, tempChain, l.UTXO, l.ValidationServerURL)

	if isValid {
		// Clear Mempool of confirmed transactions (transactions that are now in this block)
		l.MemPool = RemoveConfirmedTransactions(l.MemPool, block.Transactions)

		// Update UTXO
		l.UTXO = newUTXO

		// Update chain
		l.Chain = tempChain

		return true
	} else {
		return false
	}
}

// Takes a slice of chains and finds the longest, valid chain and sets our chain to that chain.
// It will terminate if no chains are valid or once it finds a chain smaller than our current chain. It has side effects:
//  - It removes the transactions inside the chain's blocks from the MemPool
//  - It updates the UTXO
func (l *LocalNode) Consensus(chains ...[]Block) bool {
	// Sort the changes by longest first
	sort.Slice(chains, func(index1, index2 int) bool {
		return len(chains[index1]) < len(chains[index2])
	})

	for _, chain := range chains {
		// If the chain is smaller than our current chain, our chain was the longest, so stop.
		if len(chain) < len(l.Chain) {
			fmt.Println("Our chain is longest, so our consensus function terminated.")
			return false
		}

		if valid, utxo := ValidateChain(chain, l.ValidationServerURL); valid == true {
			l.Chain = chain
			l.UTXO = utxo

			// Clear the MemPool of any confirmed transactions
			for _, block := range chain {
				l.MemPool = RemoveConfirmedTransactions(l.MemPool, block.Transactions)
			}

			// We found a longer, valid chain.
			fmt.Println("We found a valid chain through our consensus function!")
			return true
		}
	}

	// No chain was valid or chosen.
	fmt.Println("We ran our consensus function but all chains were invalid.")
	return false
}

// Finds a valid proof for a block and validates transactions from the MemPool. It removes invalid transactions.
// It returns a pointer to a new block that will be nil if the mining process was canceled.
// It does not add this block to the chain itself.
func (l LocalNode) MineBlock(shouldMine *bool) *Block {
	// Ensure that we are mining
	*shouldMine = true

	// Add a "coinbase" transaction that mints the correct amount of coins to the miner (this node's public key)
	newTransactions := append(l.MemPool, Transaction{"0", l.OperatorPublicKey, coinbaseReward, 0, ""})

	// Sort the transactions by their timestamp
	sort.Slice(newTransactions, func(index1, index2 int) bool {
		return newTransactions[index1].Timestamp < newTransactions[index2].Timestamp
	})

	// Remove invalid transactions
	for index, transaction := range newTransactions {
		// If the transaction is a coinbase transaction (the first transaction):
		if index == 0 {
			// Skip validation.
			continue
		}

		// If the sender has enough coins, and the signature is valid
		if transaction.Amount < l.UTXO[transaction.Sender] && ValidateSignature(transaction, l.ValidationServerURL) {
			// Update the balances of both parties
			l.UTXO[transaction.Sender] -= transaction.Amount
			l.UTXO[transaction.Recipient] += transaction.Amount
		} else {
			newTransactions = RemoveFromTransactions(newTransactions, index)
		}
	}

	blockHeader := BlockHeader{time.Now().Unix(), newTransactions, LastBlock(l.Chain).hash()}

	// Create a proof with the appropriate difficulty
	proof := Proof{Nonce: 0, DifficultyThreshold: DetermineDifficultyForChainIndex(l.Chain, len(l.Chain))}

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

// Runs the ValidateBlock function on each block in the chain (except the genesis block), and checks that the genesis block has not changed.
// It returns whether the chain is valid and an updated UTXO (or nil if not valid).
func ValidateChain(blocks []Block, validationServerURL string) (bool, UTXO) {
	utxo := make(UTXO)

	// Iterate over all blocks and check if they are valid (and update UTXO)
	for index, _ := range blocks {

		valid, newUTXO := ValidateBlock(index, blocks, utxo, validationServerURL)

		if !valid {
			return false, nil
		} else {
			utxo = newUTXO
		}
	}

	return true, utxo
}

// ValidateBlock takes the index of a block, the full Blockchain, a UTXO of the Blockchain up to that point, and a validationServerURL.
// It returns whether that block is valid and an updated UTXO including that block's transactions.
// Does these checks to ensure the chain is valid:
//  - Check that previous hashes are valid
//  - Check that users have enough UTXO to afford transactions
//  - Check that proofs are valid
//  - Check that there are not more than one coinbase transaction in each block
//  - Check that signatures are valid
//  - Check that difficulty threshold is valid
func ValidateBlock(index int, blocks []Block, utxo UTXO, validationServerURL string) (bool, UTXO) {
	block := blocks[index]

	// If the block is the genesis block:
	if index == 0 {
		// Check this is the correct genesis block
		if reflect.DeepEqual(block, GenesisBlock) {
			genesisTransaction := block.Transactions[0]

			utxo[genesisTransaction.Recipient] += genesisTransaction.Amount

			return true, utxo
		} else {

			// The genesis block has been tampered with! This is an invalid block!
			return false, nil
		}
	}

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
		// If the transaction is a coinbase transaction (the first transaction):
		if index == 0 {
			// If this is a VALID coinbase transaction
			if transaction.Sender == "0" && transaction.Amount == coinbaseReward {
				// Add coins to the recipient without taking from the sender (as this is a coinbase transaction)
				utxo[transaction.Recipient] += transaction.Amount
			} else {
				return false, nil
			}

			// Skip other validation
			continue
		}

		// If the sender has enough coins, and the signature is valid
		if transaction.Amount < utxo[transaction.Sender] && ValidateSignature(transaction, validationServerURL) {
			// Update the balances of both parties
			utxo[transaction.Sender] -= transaction.Amount
			utxo[transaction.Recipient] += transaction.Amount
		} else {
			return false, nil
		}

	}

	return true, utxo
}
