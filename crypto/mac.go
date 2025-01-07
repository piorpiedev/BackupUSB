package crypto

import (
	"crypto/hmac"
	"hash"

	"lukechampine.com/blake3"
)

const MACSUM_SIZE = 64 // Encrypted with AES, so EncryptedSize = Size

func NewMAC(key []byte) hash.Hash {
	return blake3.New(64, key)
}

func CompareMacSums(macSum1, macSum2 []byte) bool {
	return hmac.Equal(macSum1, macSum2) // Unnecessary, but let's leave it as is for good practice
}
