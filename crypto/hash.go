package crypto

import (
	"encoding/hex"
	"hash"
	"math/rand"

	"golang.org/x/crypto/sha3"
)

var (
	// HashStrategy is the hash strategy for SHA3-256.
	HashStrategy func() hash.Hash = sha3.New256
)

// GetSHA256Hash generates a SHA3-256 hash.
func GetSHA256Hash(b []byte) ([]byte, error) {
	h := HashStrategy()

	if _, err := h.Write(b); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// ByteArrayToHex converts a set of bytes to a hex encoded string.
func ByteArrayToHex(payload []byte) string {
	return hex.EncodeToString(payload)
}

// GenRandomBytes generates a byte array containing "n" random bytes.
func GenRandomBytes(n, seed int64) ([]byte, error) {
	randBytes := make([]byte, n)
	rand.Seed(seed)
	_, err := rand.Read(randBytes)

	return randBytes, err
}
