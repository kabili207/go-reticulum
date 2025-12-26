package cryptography

import (
	"bytes"
	"testing"
)

// RFC 5869 - Test Case 1 (HKDF-SHA256)
func TestHKDF_RFC5869_TestCase1(t *testing.T) {
	maybeParallel(t)

	ikm := bytes.Repeat([]byte{0x0b}, 22)
	salt := mustHex(t, "000102030405060708090a0b0c")
	info := mustHex(t, "f0f1f2f3f4f5f6f7f8f9")
	l := 42

	// Reference OKM from RFC5869
	wantOKM := mustHex(t, "3cb25f25faacd57a90434f64d0362f2a2d2d0a90cf1a5a4c5db02d56ecc4c5bf34007208d5b887185865")

	// hkdf.go implementation
	got, err := HKDF(l, ikm, salt, info)
	if err != nil {
		t.Fatalf("HKDF: %v", err)
	}
	if !bytes.Equal(got, wantOKM) {
		t.Fatalf("HKDF mismatch")
	}

	// crypto.go wrapper implementation
	got2, err := HKDFSHA256(salt, ikm, info, l)
	if err != nil {
		t.Fatalf("HKDFSHA256: %v", err)
	}
	if !bytes.Equal(got2, wantOKM) {
		t.Fatalf("HKDFSHA256 mismatch")
	}
}

func TestHKDF_Errors(t *testing.T) {
	maybeParallel(t)

	if _, err := HKDF(0, []byte{1}, nil, nil); err == nil {
		t.Fatalf("expected error for length=0")
	}
	if _, err := HKDF(32, nil, nil, nil); err == nil {
		t.Fatalf("expected error for empty ikm")
	}

	if _, err := HKDFSHA256(nil, []byte("ikm"), nil, 0); err == nil {
		t.Fatalf("expected error for invalid length")
	}
	// 255 blocks max
	if _, err := HKDFSHA256(nil, []byte("ikm"), nil, 32*256); err == nil {
		t.Fatalf("expected error for too-large length")
	}
}

