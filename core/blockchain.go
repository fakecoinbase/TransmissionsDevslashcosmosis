package core

import (
	log "github.com/sirupsen/logrus"
	"reflect"
	"sort"
	"time"
)

// Adds a transaction to the MemPool (but will do nothing to incorporate it into a block or verify it).
func (l *LocalNode) AddTransactionToMemPool(transaction Transaction, doNotBroadcast ...bool) {
	//TODO: If performance becomes a problem run this in a separate goroutine

	// Don't accept transactions with invalid signatures
	if !ValidateSignature(transaction, l.ValidationServerURL) {
		log.Warn("We just got a transaction with an invalid signature. It was not added.")
		return
	}

	// If transaction is not already in MemPool/Chain:
	if IsTransactionAlreadyInMemPoolOrChain(transaction, l.MemPool, l.Chain) {
		log.Warn("We just got a duplicate transaction. It was not added.")
		return
	}

	// Add transaction to MemPool.
	l.MemPool = append(l.MemPool, transaction)

	// Only broadcast if we aren't passed a doNotBroadcast param
	if len(doNotBroadcast) == 0 {
		l.BroadcastTransaction(transaction)
	}

	log.Info("We just got a new transaction!")

}

// Adds a new block to the chain (by first verifying it and getting its UTXO). It has side effects:
//  - It stops all mining processes on this node
//  - It removes the transactions inside the block from the MemPool
//  - It updates the UTXO
func (l *LocalNode) AddMinedBlockToChain(block Block, alternativePeerConsensusFunction ...func()) bool {
	// Cancel mining processes as a new block has been found
	l.IsMining = false

	// If the previous hash is not the previous block's hash:
	if block.PreviousHash != LastBlock(l.Chain).hash() {
		// We might have missed a previous block that was broadcast to us.

		// The else is only for tests. By default, only the success case will run.
		if len(alternativePeerConsensusFunction) == 0 {
			// We'll run peer consensus to get the missing block.
			l.GetPeerConsensus()
		} else {
			// Run the alternative peer consensus function
			alternativePeerConsensusFunction[0]()
		}
	}

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
		return len(chains[index1]) > len(chains[index2])
	})

	for _, chain := range chains {
		// If the chain is smaller than our current chain, our chain was the longest, so stop.
		if len(chain) < len(l.Chain) {
			log.Info("Our chain is longest, so our consensus function terminated.")
			return false
		}

		if valid, utxo := ValidateChain(chain, l.ValidationServerURL); valid == true {
			l.Chain = chain
			l.UTXO = utxo

			// Clear the MemPool of any confirmed transactions
			for _, block := range chain {
				l.MemPool = RemoveConfirmedTransactions(l.MemPool, block.Transactions)
			}

			// Cancel mining
			l.IsMining = false

			// We found a longer, valid chain.
			log.Info("We found a valid chain through our consensus function!")
			return true
		}
	}

	// No chain was valid or chosen.
	log.Warn("We ran our consensus function but all chains were invalid.")
	return false
}

// Finds a valid proof for a block and validates transactions from the MemPool. It removes invalid transactions.
// It returns a pointer to a new block that will be nil if the mining process was canceled.
// It does not add this block to the chain itself.
func (l LocalNode) MineBlock(shouldMine *bool) *Block {
	// Ensure that we are mining
	*shouldMine = true

	// Make copy of UTXO
	newUTXO := make(UTXO)
	for k, v := range l.UTXO {
		newUTXO[k] = v
	}

	// Create a copy of the MemPool
	memPool := make([]Transaction, len(l.MemPool))
	copy(memPool, l.MemPool)

	// Sort the MemPool by each transaction's timestamp
	sort.Slice(memPool, func(index1, index2 int) bool {
		return memPool[index1].Timestamp < memPool[index2].Timestamp
	})

	// Create a newTransactions slice and prepend a "coinbase" transaction that mints the correct amount of coins to the miner (this node's public key)
	newTransactions := []Transaction{{"0", l.OperatorPublicKey, coinbaseReward, time.Now().Unix(), ""}}

	// Add all valid memPool transactions to the newTransactions slice
	for _, transaction := range memPool {
		// If the transaction is valid
		if ValidateTransaction(transaction, newUTXO, l.ValidationServerURL) {
			// Update the balances of both parties
			newUTXO[transaction.Sender] -= transaction.Amount
			newUTXO[transaction.Recipient] += transaction.Amount
			// Add transaction to block's newTransactions
			newTransactions = append(newTransactions, transaction)
		}
	}

	// Don't mine if there's only one transaction (the coinbase transaction)
	if len(newTransactions) == 1 {
		log.Warn("There was only one transaction (the coinbase transaction) in a block we started mining. Canceling...")
		*shouldMine = false
		return nil
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
//  - Check that there are not duplicate transactions in the block that appear earlier in the chain
func ValidateBlock(blockIndex int, blocks []Block, utxo UTXO, validationServerURL string, shouldUseAltGenesisBlock ...bool) (bool, UTXO) {
	block := blocks[blockIndex]

	// If the block is the genesis block:
	if blockIndex == 0 {
		// We do this for tests, as tests use a separate genesis block
		// because we hardcoded signatures which will break if we use the new genesis block.
		genesisBlock := GenesisBlock
		if len(shouldUseAltGenesisBlock) == 0 {
			genesisBlock = testGenesisBlock
		}

		// Check this is the correct genesis block
		if reflect.DeepEqual(block, genesisBlock) {
			genesisTransaction := block.Transactions[0]

			utxo[genesisTransaction.Recipient] += genesisTransaction.Amount

			return true, utxo
		} else {
			// The genesis block has been tampered with! This is an invalid block!
			return false, nil
		}
	}

	// Invalid if there's only one transaction (the coinbase transaction)
	if len(block.Transactions) == 1 {
		return false, nil
	}

	// Check that difficulty threshold is valid
	if block.Proof.DifficultyThreshold != DetermineDifficultyForChainIndex(blocks, blockIndex) {
		return false, nil
	}

	lastBlock := blocks[blockIndex-1]

	// Check previous hash is valid and that proof is valid
	if block.PreviousHash != lastBlock.hash() || !ValidateProof(block) {
		return false, nil
	}

	// Check the transactions in it are valid
	for transactionIndex, transaction := range block.Transactions {
		// If the transaction is a coinbase transaction (the first transaction):
		if transactionIndex == 0 {
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

		// If the transaction is valid
		if ValidateTransaction(transaction, utxo, validationServerURL) {
			// Update the balances of both parties
			utxo[transaction.Sender] -= transaction.Amount
			utxo[transaction.Recipient] += transaction.Amount
		} else {
			return false, nil
		}

		// Check that the transaction hasn't been made previously
		if IsTransactionInChain(transaction, blocks[:blockIndex]) {
			return false, nil
		}

	}

	return true, utxo
}

// Checks if a transaction is a positive number, the sender has enough coins the make the transaction, and that the signature is valid.
func ValidateTransaction(transaction Transaction, utxo UTXO, validationServerURL string) bool {
	return transaction.Amount > 0 && transaction.Amount <= utxo[transaction.Sender] && ValidateSignature(transaction, validationServerURL)
}
