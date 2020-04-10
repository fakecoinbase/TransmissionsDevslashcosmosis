package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateSignature(t *testing.T) {
	// Valid URL
	transaction := Transaction{Sender: "0458adabe2c014de6c3fd2f2c865c2ca7fe823a4131a4d22f98dcc77f1bffc8aeacf8a0b7949321c33214e9c1b2201063404a321110be8223ad1685ee32c9c02d0", Recipient: "046007e213c57ccab18af3f3b385893da75514ab691216152955d70937744dbe040de0ea504ebe29bce2476ae37c794cf5e7d96c8bc2ad153eb434b148f1af6f6c", Amount: 15, Timestamp: 1586117966, Signature: "3046022100d158259aae3c7c9e3e6cd33a3b47134723ddc4cae25484e8a5df28f45ee462fd022100b6c6600f89a3ef050a8aab14c8a96ca5b5b9c8fa358945c9f53dda1b488dd43c"}
	assert.True(t, ValidateSignature(transaction, "https://crows.sh/verifySignature"))

	invalidTransaction := Transaction{Sender: "0458adabe2c014de6c3fd2f2c865c2ca7fe823a4131a4d22f98dcc77f1bffc8aeacf8a0b7949321c33214e9c1b2201063404a321110be8223ad1685ee32c9c02d0", Recipient: "046007e213c57ccab18af3f3b385893da75514ab691216152955d70937744dbe040de0ea504ebe29bce2476ae37c794cf5e7d96c8bc2ad153eb434b148f1af6f6c", Amount: 90000000, Timestamp: 1586117966, Signature: "f5f036c0117dd360e57affe1ad76cdb7486f6befd44a8aa201a6713426dd77891ee7263ee2b62449f44ac56f1a83caf9f813727f91f0e66d3da8ed96846e8d4d"}
	assert.False(t, ValidateSignature(invalidTransaction, "https://crows.sh/verifySignature"))

	// Invalid URL
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "ValidateSignature did not panic with invalid validationServerURL!")
		}
	}()
	ValidateSignature(transaction, "notreal.google.com")
}
