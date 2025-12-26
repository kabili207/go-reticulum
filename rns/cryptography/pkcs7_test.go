package cryptography

import (
	"bytes"
	"testing"
)

func TestPKCS7PadUnpad_RoundTrip_DefaultBlockSize(t *testing.T) {
	maybeParallel(t)

	lengths := []int{0, 1, 15, 16, 17, 31, 32, 33, 63, 64}
	for _, n := range lengths {
		t.Run("len="+itoa(n), func(t *testing.T) {
			in := bytes.Repeat([]byte{0xA5}, n)
			padded, err := PKCS7Pad(in, 0)
			if err != nil {
				t.Fatalf("pad: %v", err)
			}
			if len(padded) == 0 || len(padded)%PKCS7BlockSize != 0 {
				t.Fatalf("unexpected padded length %d", len(padded))
			}

			out, err := PKCS7Unpad(padded, 0)
			if err != nil {
				t.Fatalf("unpad: %v", err)
			}
			if !bytes.Equal(out, in) {
				t.Fatalf("roundtrip mismatch")
			}
		})
	}
}

func TestPKCS7Pad_BlockSizeValidation(t *testing.T) {
	maybeParallel(t)

	if _, err := PKCS7Pad([]byte("x"), 256); err == nil {
		t.Fatalf("expected error for bs=256")
	}
	if _, err := PKCS7Unpad([]byte(""), 256); err == nil {
		t.Fatalf("expected error for bs=256")
	}
}

func TestPKCS7Unpad_Errors(t *testing.T) {
	maybeParallel(t)

	// Empty / non-multiple length
	if _, err := PKCS7Unpad(nil, 16); err == nil {
		t.Fatalf("expected error for nil input")
	}
	if _, err := PKCS7Unpad([]byte{}, 16); err == nil {
		t.Fatalf("expected error for empty input")
	}
	if _, err := PKCS7Unpad([]byte("123"), 16); err == nil {
		t.Fatalf("expected error for non-multiple length")
	}

	// Bad padLen 0
	bad0 := bytes.Repeat([]byte{0}, 16)
	bad0[15] = 0
	if _, err := PKCS7Unpad(bad0, 16); err == nil {
		t.Fatalf("expected error for padLen=0")
	}

	// Bad padLen > bs
	badBig := bytes.Repeat([]byte{0}, 16)
	badBig[15] = 17
	if _, err := PKCS7Unpad(badBig, 16); err == nil {
		t.Fatalf("expected error for padLen>bs")
	}

	// Bad padding bytes
	badBytes := bytes.Repeat([]byte{0}, 16)
	badBytes[15] = 2
	badBytes[14] = 3
	if _, err := PKCS7Unpad(badBytes, 16); err == nil {
		t.Fatalf("expected error for invalid padding bytes")
	}
}

func itoa(n int) string {
	// small helper to avoid fmt in tight test loops
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [32]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

