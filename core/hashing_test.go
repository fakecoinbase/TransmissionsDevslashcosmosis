package core

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSHA256(t *testing.T) {
	// Test SHA256
	hashed := SHA256("test")
	assert.Equal(t, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", hashed)
	hashed2 := SHA256("test.")
	assert.Equal(t, "4ee3df88f682d376531d8803f2ccbee56d075cd248fc300f55dfe8596a7354b7", hashed2)

	// Test block.hash()
	block := Block{BlockHeader: BlockHeader{Timestamp: 1586119312, Transactions: []Transaction{Transaction{Sender: "0", Recipient: "0458adabe2c014de6c3fd2f2c865c2ca7fe823a4131a4d22f98dcc77f1bffc8aeacf8a0b7949321c33214e9c1b2201063404a321110be8223ad1685ee32c9c02d0", Amount: 1000, Timestamp: 0, Signature: ""}, Transaction{Sender: "0458adabe2c014de6c3fd2f2c865c2ca7fe823a4131a4d22f98dcc77f1bffc8aeacf8a0b7949321c33214e9c1b2201063404a321110be8223ad1685ee32c9c02d0", Recipient: "6007e213c57ccab18af3f3b385893da75514ab691216152955d70937744dbe040de0ea504ebe29bce2476ae37c794cf5e7d96c8bc2ad153eb434b148f1af6f6c", Amount: 15, Timestamp: 1586117966, Signature: "f5f036c0117dd360e57affe1ad76cdb7486f6befd44a8aa201a6713426dd77891ee7263ee2b62449f44ac56f1a83caf9f813727f91f0e66d3da8ed96846e8d4d"}}, PreviousHash: "b83312421b34ba8bc36351d52df47abb6f3c9284897f890fdece2b561859eeb5"}, Proof: Proof{Nonce: 659410, DifficultyThreshold: 5}}
	assert.Equal(t, SHA256(fmt.Sprintf("%v", block)), block.hash())
}
