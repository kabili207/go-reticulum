package cryptography

import (
	"bytes"
	crand "crypto/rand"
	"testing"
)

func TestToken_EncryptDecrypt_RoundTrip(t *testing.T) {
	maybeParallel(t)

	key, err := GenerateKey(32)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	tok, err := NewToken(key)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	plain := []byte("hello token")
	enc, err := tok.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	out, err := tok.Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("roundtrip mismatch")
	}
}

func TestToken_TamperDetection(t *testing.T) {
	maybeParallel(t)

	key := bytes.Repeat([]byte{0x42}, 64)
	tok, err := NewToken(key)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	enc, err := tok.Encrypt([]byte("hello"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	flip := func(i int) []byte {
		cp := append([]byte(nil), enc...)
		cp[i] ^= 0x01
		return cp
	}

	// Flip IV byte.
	if _, err := tok.Decrypt(flip(0)); err == nil {
		t.Fatalf("expected HMAC error on IV tamper")
	}
	// Flip ciphertext byte.
	if _, err := tok.Decrypt(flip(blockSize)); err == nil {
		t.Fatalf("expected HMAC error on ciphertext tamper")
	}
	// Flip HMAC byte.
	if _, err := tok.Decrypt(flip(len(enc) - 1)); err == nil {
		t.Fatalf("expected HMAC error on signature tamper")
	}
}

func TestToken_Errors(t *testing.T) {
	maybeParallel(t)

	if _, err := NewToken(nil); err == nil {
		t.Fatalf("expected error for nil key")
	}
	if _, err := NewToken(bytes.Repeat([]byte{0x00}, 10)); err == nil {
		t.Fatalf("expected error for wrong key length")
	}

	key := bytes.Repeat([]byte{0x11}, 64)
	tok, _ := NewToken(key)

	if _, err := tok.Encrypt(nil); err == nil {
		t.Fatalf("expected error for nil plaintext")
	}
	if _, err := tok.Decrypt(nil); err == nil {
		t.Fatalf("expected error for nil token")
	}
	if _, err := tok.Decrypt([]byte{1, 2, 3}); err == nil {
		t.Fatalf("expected error for too-short token")
	}
}

func TestToken_Decrypt_CiphertextLengthCheckAfterValidHMAC(t *testing.T) {
	// Do not run in parallel: modifies crypto/rand.Reader.

	// Build a token with a deterministic IV to keep the test stable.
	orig := crand.Reader
	defer func() { crand.Reader = orig }()
	crand.Reader = bytes.NewReader(bytes.Repeat([]byte{0xAA}, 64))

	key := bytes.Repeat([]byte{0x22}, 64)
	tok, err := NewToken(key)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	// Create a valid token, then truncate ciphertext length by 1 and recompute HMAC.
	enc, err := tok.Encrypt([]byte("hello"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	iv := enc[:blockSize]
	ciphertext := enc[blockSize : len(enc)-hmacSize]
	if len(ciphertext) < blockSize {
		t.Fatalf("unexpected ciphertext length %d", len(ciphertext))
	}
	ciphertext = ciphertext[:len(ciphertext)-1] // break block alignment

	// Recompute valid HMAC for malformed body.
	body := append(append([]byte(nil), iv...), ciphertext...)
	mac := tok.signingKey // tests are in-package, ok to access
	bad := append(append([]byte(nil), body...), computeHMACSHA256(mac, body)...)

	if _, err := tok.Decrypt(bad); err == nil {
		t.Fatalf("expected error for non-block-multiple ciphertext")
	}
}

func computeHMACSHA256(key, data []byte) []byte {
	h := NewHMAC(key, data, nil)
	sum := h.Digest()
	return sum[:hmacSize]
}

