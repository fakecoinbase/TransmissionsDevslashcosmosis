package core

// Gets the most recent link in a chain of blocks.
func LastBlock(chain []Block) Block {
	return chain[len(chain)-1]
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

// Checks if a transaction is in a list of confirmed transactions.
func isTransactionConfirmed(transaction Transaction, confirmedTransactions []Transaction) bool {
	for _, confirmedTransaction := range confirmedTransactions {
		if transaction == confirmedTransaction {
			return true
		}
	}

	return false
}

// Calculates the mean of a slice.
func calcMean(input []float64) float64 {
	total := 0.0

	for _, v := range input {
		total += v
	}

	return total / float64(len(input))
}
