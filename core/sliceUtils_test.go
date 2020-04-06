package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsTransactionInMemPool(t *testing.T) {
	there := IsTransactionInMemPool(Transaction{Signature: "test1"}, []Transaction{{Signature: "test2"}, {Signature: "test1"}, {Signature: "test3"}})
	notThere := IsTransactionInMemPool(Transaction{Signature: "test1"}, []Transaction{{Signature: "test2"}, {Signature: "test3"}})

	assert.True(t, there)

	assert.False(t, notThere)

}

func TestIsTransactionInChain(t *testing.T) {
	there := IsTransactionInChain(Transaction{Signature: "test1"}, []Block{{BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test1"}}}}, {BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test2"}}}}, {BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test4"}}}}})
	notThere := IsTransactionInChain(Transaction{Signature: "test1"}, []Block{{BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test2"}}}}, {BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test3"}}}}})

	assert.True(t, there)

	assert.False(t, notThere)
}

func TestIsTransactionAlreadyInMemPoolOrChain(t *testing.T) {
	inMemPoolAndChain := IsTransactionAlreadyInMemPoolOrChain(Transaction{Signature: "test1"}, []Transaction{{Signature: "test2"}, {Signature: "test1"}, {Signature: "test3"}}, []Block{{BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test1"}}}}, {BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test2"}}}}, {BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test4"}}}}})
	inChain := IsTransactionAlreadyInMemPoolOrChain(Transaction{Signature: "test1"}, []Transaction{{Signature: "test2"}, {Signature: "test3"}}, []Block{{BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test1"}}}}, {BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test3"}}}}})
	inMemPool := IsTransactionAlreadyInMemPoolOrChain(Transaction{Signature: "test1"}, []Transaction{{Signature: "test1"}, {Signature: "test3"}}, []Block{{BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test0"}}}}, {BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test3"}}}}})
	inNeither := IsTransactionAlreadyInMemPoolOrChain(Transaction{Signature: "NOPE"}, []Transaction{{Signature: "test1"}, {Signature: "test3"}}, []Block{{BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test0"}}}}, {BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test3"}}}}})

	assert.True(t, inMemPoolAndChain)
	assert.True(t, inChain)
	assert.True(t, inMemPool)
	assert.False(t, inNeither)
}

func TestLastBlock(t *testing.T) {
	lastBlock := LastBlock([]Block{{BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test2"}}}}, {BlockHeader: BlockHeader{Transactions: []Transaction{{Signature: "test3"}}}}})
	assert.Equal(t, lastBlock.Transactions[0].Signature, "test3")
}

func TestRemoveFromTransactions(t *testing.T) {
	filtered := RemoveFromTransactions([]Transaction{{Signature: "test2"}, {Signature: "test1"}, {Signature: "test3"}}, 0)

	assert.NotContains(t, filtered, Transaction{Signature: "test2"})
}

func TestRemoveConfirmedTransactions(t *testing.T) {
	filtered := RemoveConfirmedTransactions([]Transaction{{Signature: "test2"}, {Signature: "test1"}, {Signature: "test3"}}, []Transaction{{Signature: "test2"}})
	assert.NotContains(t, filtered, Transaction{Signature: "test2"})
}

func TestIsTransactionConfirmed(t *testing.T) {
	isConfirmed := isTransactionConfirmed(Transaction{Signature: "test2"}, []Transaction{{Signature: "test2"}, {Signature: "test1"}, {Signature: "test3"}})
	assert.True(t, isConfirmed)
}

func TestCalcMean(t *testing.T) {
	mean := calcMean([]float64{0.0, 5.0, 10.0})
	assert.Equal(t, mean, 5.0)
}
