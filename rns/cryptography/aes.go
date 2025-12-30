package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

const BlockSize = aes.BlockSize // 16

// AES128CBCEncrypt/AES128CBCDecrypt are equivalent to Python AES_128_CBC.
func AES128CBCEncrypt(plaintext, key, iv []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("invalid AES key length"), len(key))
	}
	if len(iv) != BlockSize {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("invalid AES IV length"), len(iv))
	}
	if len(plaintext)%BlockSize != 0 {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("data length must be multiple of AES block size (16)"), len(plaintext))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return ciphertext, nil
}

func AES128CBCDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("invalid AES key length"), len(key))
	}
	if len(iv) != BlockSize {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("invalid AES IV length"), len(iv))
	}
	if len(ciphertext)%BlockSize != 0 {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("data length must be multiple of AES block size (16)"), len(ciphertext))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	return plaintext, nil
}

// AES256CBCEncrypt/AES256CBCDecrypt are equivalent to Python AES_256_CBC.
func AES256CBCEncrypt(plaintext, key, iv []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("invalid AES key length"), len(key))
	}
	if len(iv) != BlockSize {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("invalid AES IV length"), len(iv))
	}
	if len(plaintext)%BlockSize != 0 {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("data length must be multiple of AES block size (16)"), len(plaintext))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return ciphertext, nil
}

func AES256CBCDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("invalid AES key length"), len(key))
	}
	if len(iv) != BlockSize {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("invalid AES IV length"), len(iv))
	}
	if len(ciphertext)%BlockSize != 0 {
		return nil, fmt.Errorf("%w: %d bytes", fmt.Errorf("data length must be multiple of AES block size (16)"), len(ciphertext))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	return plaintext, nil
}
