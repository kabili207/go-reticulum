package cryptography

import (
	"crypto/sha256"
	"crypto/sha512"
)

// sha256(data) → returns 32-byte hash
func Sha256(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

// sha512(data) → returns 64-byte hash
func Sha512(data []byte) []byte {
	h := sha512.Sum512(data)
	return h[:]
}
