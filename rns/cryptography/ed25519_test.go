package cryptography

import (
	"bytes"
	"testing"
)

func TestEd25519_SignVerify(t *testing.T) {
	maybeParallel(t)

	seed := bytes.Repeat([]byte{0x01}, 32)
	priv, err := NewEd25519PrivateKey(seed)
	if err != nil {
		t.Fatalf("NewEd25519PrivateKey: %v", err)
	}
	pub := priv.PublicKey()

	msg := []byte("hello")
	sig := priv.Sign(msg)

	if err := pub.Verify(sig, msg); err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if err := pub.Verify(sig, []byte("hello!")); err == nil {
		t.Fatalf("expected verify failure for modified message")
	}
}

func TestEd25519_KeyLengthErrors(t *testing.T) {
	maybeParallel(t)

	if _, err := NewEd25519PrivateKey([]byte("short")); err == nil {
		t.Fatalf("expected seed length error")
	}
	if _, err := NewEd25519PublicKey([]byte("short")); err == nil {
		t.Fatalf("expected public key length error")
	}
}

