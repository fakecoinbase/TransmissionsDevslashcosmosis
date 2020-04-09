package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateProof(t *testing.T) {
	invalidProof := ValidateProof(Block{BlockHeader: BlockHeader{Timestamp: 0, Transactions: []Transaction{Transaction{Sender: "0", Recipient: "0458adabe2c014de6c3fd2f2c865c2ca7fe823a4131a4d22f98dcc77f1bffc8aeacf8a0b7949321c33214e9c1b2201063404a321110be8223ad1685ee32c9c02d0", Amount: 1000, Timestamp: 0, Signature: ""}, Transaction{Sender: "0458adabe2c014de6c3fd2f2c865c2ca7fe823a4131a4d22f98dcc77f1bffc8aeacf8a0b7949321c33214e9c1b2201063404a321110be8223ad1685ee32c9c02d0", Recipient: "6007e213c57ccab18af3f3b385893da75514ab691216152955d70937744dbe040de0ea504ebe29bce2476ae37c794cf5e7d96c8bc2ad153eb434b148f1af6f6c", Amount: 15, Timestamp: 1586117966, Signature: "f5f036c0117dd360e57affe1ad76cdb7486f6befd44a8aa201a6713426dd77891ee7263ee2b62449f44ac56f1a83caf9f813727f91f0e66d3da8ed96846e8d4d"}}, PreviousHash: "b83312421b34ba8bc36351d52df47abb6f3c9284897f890fdece2b561859eeb5"}, Proof: Proof{Nonce: 659410, DifficultyThreshold: 5}})
	validProof := ValidateProof(Block{BlockHeader: BlockHeader{Timestamp: 1586119312, Transactions: []Transaction{Transaction{Sender: "0", Recipient: "0458adabe2c014de6c3fd2f2c865c2ca7fe823a4131a4d22f98dcc77f1bffc8aeacf8a0b7949321c33214e9c1b2201063404a321110be8223ad1685ee32c9c02d0", Amount: 1000, Timestamp: 0, Signature: ""}, Transaction{Sender: "0458adabe2c014de6c3fd2f2c865c2ca7fe823a4131a4d22f98dcc77f1bffc8aeacf8a0b7949321c33214e9c1b2201063404a321110be8223ad1685ee32c9c02d0", Recipient: "6007e213c57ccab18af3f3b385893da75514ab691216152955d70937744dbe040de0ea504ebe29bce2476ae37c794cf5e7d96c8bc2ad153eb434b148f1af6f6c", Amount: 15, Timestamp: 1586117966, Signature: "f5f036c0117dd360e57affe1ad76cdb7486f6befd44a8aa201a6713426dd77891ee7263ee2b62449f44ac56f1a83caf9f813727f91f0e66d3da8ed96846e8d4d"}}, PreviousHash: "b83312421b34ba8bc36351d52df47abb6f3c9284897f890fdece2b561859eeb5"}, Proof: Proof{Nonce: 659410, DifficultyThreshold: 5}})

	assert.False(t, invalidProof)
	assert.True(t, validProof)
}

func TestDetermineDifficultyForChainIndex(t *testing.T) {
	chain := []Block{{BlockHeader: BlockHeader{Timestamp: 0}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 600}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 1200}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 1800}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 2400}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 3000}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 3600}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 4200}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 4800}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 5400}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 6000}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 6600}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 7200}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 7800}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 8400}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 9000}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 9600}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 10200}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 10800}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 11400}, Proof: Proof{DifficultyThreshold: 10}}}

	assert.Equal(t, int64(10), DetermineDifficultyForChainIndex(chain, 20))

	chain2 := []Block{{BlockHeader: BlockHeader{Timestamp: 0}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 600}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 1200}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 1800}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 2400}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 3000}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 3600}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 4200}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 4800}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 5400}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 6000}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 6600}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 7200}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 7800}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 8400}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 9000}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 9600}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 10200}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 10800}, Proof: Proof{DifficultyThreshold: 10}}, {BlockHeader: BlockHeader{Timestamp: 11000}, Proof: Proof{DifficultyThreshold: 10}}}
	assert.Equal(t, int64(11), DetermineDifficultyForChainIndex(chain2, 20))

	// Index is less than 10 (meaning there aren't enough past blocks to examine), just returning a difficulty of 5.
	assert.Equal(t, int64(5), DetermineDifficultyForChainIndex(chain2, 9))
}
