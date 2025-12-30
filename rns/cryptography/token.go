package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

const (
	Overhead   = 48 // 16 IV + 32 HMAC
	blockSize  = aes.BlockSize
	hmacSize   = 32
	key128Size = 32 // 16+16
	key256Size = 64 // 32+32
)

// GenerateKey mirrors Token.generate_key(mode=AES_256_CBC by default).
// aesKeyBytes = 16 -> AES-128 (final key 32 bytes)
// aesKeyBytes = 32 -> AES-256 (final key 64 bytes)
func GenerateKey(aesKeyBytes int) ([]byte, error) {
	var total int
	switch aesKeyBytes {
	case 16:
		total = key128Size
	case 32:
		total = key256Size
	default:
		return nil, fmt.Errorf("invalid AES key size: %d", aesKeyBytes)
	}

	k := make([]byte, total)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil, err
	}
	return k, nil
}

type Token struct {
	signingKey    []byte
	encryptionKey []byte
	aesKeySize    int // 16 or 32
}

// NewToken mirrors __init__(key, mode=AES).
func NewToken(key []byte) (*Token, error) {
	if key == nil {
		return nil, errors.New("token key cannot be nil")
	}

	switch len(key) {
	case key128Size:
		return &Token{
			signingKey:    key[:16],
			encryptionKey: key[16:],
			aesKeySize:    16,
		}, nil
	case key256Size:
		return &Token{
			signingKey:    key[:32],
			encryptionKey: key[32:],
			aesKeySize:    32,
		}, nil
	default:
		return nil, fmt.Errorf("token key must be 128 or 256 bits, not %d", len(key)*8)
	}
}

// verifyHMAC mirrors verify_hmac().
func (t *Token) verifyHMAC(tok []byte) bool {
	if len(tok) <= hmacSize {
		return false
	}

	data := tok[:len(tok)-hmacSize]
	rec := tok[len(tok)-hmacSize:]

	mac := hmac.New(sha256.New, t.signingKey)
	_, _ = mac.Write(data)
	exp := mac.Sum(nil)[:hmacSize]

	return hmac.Equal(rec, exp)
}

// Encrypt mirrors encrypt().
func (t *Token) Encrypt(data []byte) ([]byte, error) {
	if data == nil {
		return nil, errors.New("plaintext cannot be nil")
	}

	iv := make([]byte, blockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	padded, err := PKCS7Pad(data, blockSize)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(t.encryptionKey)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	signedParts := append(iv, ciphertext...)

	mac := hmac.New(sha256.New, t.signingKey)
	_, _ = mac.Write(signedParts)
	sig := mac.Sum(nil)[:hmacSize]

	return append(signedParts, sig...), nil
}

// Decrypt mirrors decrypt().
func (t *Token) Decrypt(tok []byte) ([]byte, error) {
	if tok == nil {
		return nil, errors.New("token cannot be nil")
	}
	if len(tok) < blockSize+hmacSize {
		return nil, fmt.Errorf("token too short: %d bytes", len(tok))
	}
	if !t.verifyHMAC(tok) {
		return nil, errors.New("token HMAC was invalid")
	}

	iv := tok[:blockSize]
	ciphertext := tok[blockSize : len(tok)-hmacSize]

	if len(ciphertext)%blockSize != 0 {
		return nil, fmt.Errorf("ciphertext length not multiple of block size")
	}

	block, err := aes.NewCipher(t.encryptionKey)
	if err != nil {
		return nil, err
	}

	plaintextPadded := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintextPadded, ciphertext)

	plaintext, err := PKCS7Unpad(plaintextPadded, blockSize)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt token: %w", err)
	}

	return plaintext, nil
}
