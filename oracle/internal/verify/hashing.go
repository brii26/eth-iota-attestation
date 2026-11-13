package verify

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// ComputeSHA256 returns the 32-byte hash and the lower-case hex string of raw data bytes.
func ComputeSHA256(data []byte) ([32]byte, string) {
	sum := sha256.Sum256(data)
	return sum, hex.EncodeToString(sum[:])
}

// CompareHashes checks if a claimed hash equals an expected hash in constant time.
func CompareHashes(claimed, expected string) bool {
	claimedClean := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(claimed)), "0x")
	expectedClean := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(expected)), "0x")

	if len(claimedClean) != len(expectedClean) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(claimedClean), []byte(expectedClean)) == 1
}
