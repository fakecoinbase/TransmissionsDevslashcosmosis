package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateSignature(t *testing.T) {
	// Valid URL
	transaction := Transaction{Sender: "b61e63485c4782d6495aa0091c6785d8b6c0a945a23d9b158093bbf3d93d6bb9024e6cab467cc11b51e1b1a158637a778473418298b09a7dd39c148863b1833c", Recipient: "6007e213c57ccab18af3f3b385893da75514ab691216152955d70937744dbe040de0ea504ebe29bce2476ae37c794cf5e7d96c8bc2ad153eb434b148f1af6f6c", Amount: 15, Timestamp: 1586117966, Signature: "f5f036c0117dd360e57affe1ad76cdb7486f6befd44a8aa201a6713426dd77891ee7263ee2b62449f44ac56f1a83caf9f813727f91f0e66d3da8ed96846e8d4d"}
	assert.True(t, ValidateSignature(transaction, "https://crows.sh/verifySignature"))

	invalidTransaction := Transaction{Sender: "b61e63485c4782d6495aa0091c6785d8b6c0a945a23d9b158093bbf3d93d6bb9024e6cab467cc11b51e1b1a158637a778473418298b09a7dd39c148863b1833c", Recipient: "6007e213c57ccab18af3f3b385893da75514ab691216152955d70937744dbe040de0ea504ebe29bce2476ae37c794cf5e7d96c8bc2ad153eb434b148f1af6f6c", Amount: 90000000, Timestamp: 1586117966, Signature: "f5f036c0117dd360e57affe1ad76cdb7486f6befd44a8aa201a6713426dd77891ee7263ee2b62449f44ac56f1a83caf9f813727f91f0e66d3da8ed96846e8d4d"}
	assert.False(t, ValidateSignature(invalidTransaction, "https://crows.sh/verifySignature"))

	// Invalid URL
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "ValidateSignature did not panic with invalid validationServerURL!")
		}
	}()

	ValidateSignature(transaction, "notreal.google.com")

}
