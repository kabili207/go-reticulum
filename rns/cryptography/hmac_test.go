package cryptography

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"testing"
)

func TestHMAC_MatchesStdlib_SHA256(t *testing.T) {
	maybeParallel(t)

	cases := []struct {
		name string
		key  []byte
		msg  []byte
	}{
		{"empty", []byte{}, []byte{}},
		{"short_key", []byte("k"), []byte("msg")},
		{"exact_block", bytes.Repeat([]byte{0x11}, 64), []byte("msg")},
		{"long_key", bytes.Repeat([]byte{0x22}, 200), bytes.Repeat([]byte("a"), 1000)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			std := hmac.New(sha256.New, tc.key)
			std.Write(tc.msg)
			want := std.Sum(nil)

			got := DigestFast(tc.key, tc.msg, sha256.New)
			if !bytes.Equal(got, want) {
				t.Fatalf("digest mismatch")
			}
		})
	}
}

func TestHMAC_CopyIsIndependent(t *testing.T) {
	maybeParallel(t)

	h := NewHMAC([]byte("key"), []byte("a"), sha256.New)
	c := h.Copy()

	h.Update([]byte("b"))
	c.Update([]byte("c"))

	if bytes.Equal(h.Digest(), c.Digest()) {
		t.Fatalf("expected different digests after divergent updates")
	}
}

