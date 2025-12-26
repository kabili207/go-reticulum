package cryptography

import (
	"errors"
	"fmt"
)

const PKCS7BlockSize = 16

// PKCS7Pad аналог PKCS7.pad(data, bs=BLOCKSIZE)
func PKCS7Pad(data []byte, bs int) ([]byte, error) {
	if bs <= 0 {
		bs = PKCS7BlockSize
	}
	if bs >= 256 {
		return nil, fmt.Errorf("pkcs7: invalid block size %d", bs)
	}
	padLen := bs - (len(data) % bs)
	if padLen == 0 {
		padLen = bs
	}
	out := make([]byte, len(data)+padLen)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(padLen)
	}
	return out, nil
}

// PKCS7Unpad аналог PKCS7.unpad(data, bs=BLOCKSIZE)
func PKCS7Unpad(data []byte, bs int) ([]byte, error) {
	if bs <= 0 {
		bs = PKCS7BlockSize
	}
	if bs <= 0 || bs >= 256 {
		return nil, fmt.Errorf("pkcs7: invalid block size %d", bs)
	}
	if len(data) == 0 || len(data)%bs != 0 {
		return nil, errors.New("pkcs7: invalid padded data length")
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > bs || padLen > len(data) {
		return nil, errors.New("pkcs7: invalid padding")
	}
	for i := len(data) - padLen; i < len(data); i++ {
		if data[i] != byte(padLen) {
			return nil, errors.New("pkcs7: invalid padding bytes")
		}
	}
	return data[:len(data)-padLen], nil
}
