package cryptography

import (
	"bytes"
	"testing"
)

func TestX25519_ExchangeSymmetry(t *testing.T) {
	maybeParallel(t)

	a, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	b, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	ap, err := a.PublicKey()
	if err != nil {
		t.Fatalf("PublicKey: %v", err)
	}
	bp, err := b.PublicKey()
	if err != nil {
		t.Fatalf("PublicKey: %v", err)
	}

	s1, err := a.Exchange(bp)
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	s2, err := b.Exchange(ap)
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if !bytes.Equal(s1, s2) {
		t.Fatalf("shared secret mismatch")
	}
}

func TestX25519_ClampFromPrivateBytes(t *testing.T) {
	maybeParallel(t)

	raw := bytes.Repeat([]byte{0xFF}, 32)
	k, err := FromPrivateBytes(raw)
	if err != nil {
		t.Fatalf("FromPrivateBytes: %v", err)
	}
	p := k.PrivateBytes()
	if len(p) != 32 {
		t.Fatalf("unexpected private length %d", len(p))
	}
	// clampSecret invariants
	if p[0]&7 != 0 {
		t.Fatalf("expected low bits cleared")
	}
	if p[31]&0x80 != 0 {
		t.Fatalf("expected high bit cleared")
	}
	if p[31]&0x40 == 0 {
		t.Fatalf("expected bit 6 set")
	}
}

func TestX25519_Errors(t *testing.T) {
	maybeParallel(t)

	if _, err := FromPrivateBytes([]byte("short")); err == nil {
		t.Fatalf("expected private length error")
	}
	if _, err := FromPublicBytes([]byte("short")); err == nil {
		t.Fatalf("expected public length error")
	}
	k, _ := Generate()
	if _, err := k.Exchange(123); err == nil {
		t.Fatalf("expected unsupported peer type error")
	}
}

