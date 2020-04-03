package core

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Calculate the required difficulty threshold for an index in a chain based on the past delays in timestamps.
// Will default to 5 if chain length is less than 10.
func DetermineDifficultyForChainIndex(chain []Block, index int) int64 {
	if len(chain) < 10 {
		return 5
	}

	past10Blocks := chain[index-10 : index]

	delays := make([]float64, 9)
	difficulties := make([]float64, 9)

	for i, block := range past10Blocks[1:] {

		delays = append(delays, time.Unix(block.Timestamp, 0).Sub(time.Unix(past10Blocks[i].Timestamp, 0)).Minutes())
		difficulties = append(difficulties, float64(block.Proof.DifficultyThreshold))
	}

	meanDelay := calcMean(delays)
	meanDifficulty := calcMean(difficulties)

	return int64(math.Round((10 / meanDelay) * meanDifficulty))
}

// Checks whether a block has a proof that validates it.
func ValidateProof(block Block) bool {
	// Format the proof and header
	guess := fmt.Sprintf("%v-%v", block.Proof, block.BlockHeader)
	// Hash the guess
	hashedGuess := SHA256(guess)

	// Check that the hash's first x characters are 0
	return strings.HasPrefix(hashedGuess, strings.Repeat("0", int(block.Proof.DifficultyThreshold)))
}
