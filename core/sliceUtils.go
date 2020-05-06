package core

// IsTransactionAlreadyInMemPoolOrChain checks whether the transaction already exists in a given MemPool + Blockchain.
func IsTransactionAlreadyInMemPoolOrChain(t Transaction, memPool []Transaction, chain []Block) bool {
	return IsTransactionInMemPool(t, memPool) || IsTransactionInChain(t, chain)
}

// IsTransactionAlreadyInMemPoolOrChain checks whether the transaction already exists in a given MemPool.
func IsTransactionInMemPool(t Transaction, memPool []Transaction) bool {
	for i := len(memPool) - 1; i >= 0; i-- {
		memPoolTransaction := memPool[i]

		if memPoolTransaction.Signature == t.Signature {
			return true
		}
	}

	return false
}

// IsTransactionInChain checks whether the transaction already exists in a given Blockchain.
func IsTransactionInChain(t Transaction, chain []Block) bool {
	for i := len(chain) - 1; i >= 0; i-- {

		chainBlock := chain[i]

		for _, blockTransaction := range chainBlock.Transactions {
			// If the time of the Block's transaction is more than 25 hours before the time of the incoming transaction,
			// its safe to assume that its not in the Blockchain.

			if blockTransaction.Signature == t.Signature {
				return true
			}
		}
	}

	return false
}

// Checks if a transaction is in a list of confirmed transactions.
func isTransactionConfirmed(transaction Transaction, confirmedTransactions []Transaction) bool {
	for _, confirmedTransaction := range confirmedTransactions {
		if transaction == confirmedTransaction {
			return true
		}
	}

	return false
}

// Removes a transaction from a slice at x index.
func RemoveFromTransactions(slice []Transaction, s int) []Transaction {
	return append(slice[:s], slice[s+1:]...)
}

// Takes a list of transactions and a list of transactions that have been confirmed, and removes the ones that have been confirmed.
func RemoveConfirmedTransactions(memPool []Transaction, confirmedTransactions []Transaction) []Transaction {
	filteredMemPool := make([]Transaction, 0)

	for _, transaction := range memPool {
		if !isTransactionConfirmed(transaction, confirmedTransactions) {
			filteredMemPool = append(filteredMemPool, transaction)
		}
	}

	return filteredMemPool
}

// Gets the most recent link in a chain of blocks.
func LastBlock(chain []Block) Block {
	return chain[len(chain)-1]
}

// Calculates the mean of a slice.
func calcMean(input []float64) float64 {
	total := 0.0

	for _, v := range input {
		total += v
	}

	return total / float64(len(input))
}
