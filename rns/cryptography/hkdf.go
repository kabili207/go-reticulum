package cryptography

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
)

const hkdfHashLen = 32

func hkdfHmacSHA256(key, data []byte) []byte {
	m := hmac.New(sha256.New, key)
	m.Write(data)
	return m.Sum(nil)
}

// HKDF mirrors hkdf() from HKDF.py.
func HKDF(length int, deriveFrom, salt, context []byte) ([]byte, error) {
	if length < 1 {
		return nil, errors.New("hkdf: invalid output key length")
	}
	if len(deriveFrom) == 0 {
		return nil, errors.New("hkdf: cannot derive key from empty input material")
	}

	if len(salt) == 0 {
		salt = make([]byte, hkdfHashLen)
	}
	if context == nil {
		context = []byte{}
	}

	// Extract
	pseudorandomKey := hkdfHmacSHA256(salt, deriveFrom)

	// Expand
	block := []byte{}
	derived := make([]byte, 0, length)

	nBlocks := (length + hkdfHashLen - 1) / hkdfHashLen

	for i := 0; i < nBlocks; i++ {
		mac := hmac.New(sha256.New, pseudorandomKey)
		mac.Write(block)
		mac.Write(context)
		mac.Write([]byte{byte((i + 1) % 256)}) // (i+1)%(0xFF+1) like Python
		block = mac.Sum(nil)
		derived = append(derived, block...)
	}

	return derived[:length], nil
}
