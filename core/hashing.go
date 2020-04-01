package core

import (
	"crypto/sha256"
	"fmt"
)

// Convince function that hashes a block.
func (b Block) hash() string {
	return SHA256(b)
}

// Hashes any type with SHA256 and converts to hex.
func SHA256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}
