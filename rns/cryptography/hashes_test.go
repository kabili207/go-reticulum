package cryptography

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"testing"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("invalid hex: %v", err)
	}
	return b
}

func TestSHA256_KnownVectors(t *testing.T) {
	maybeParallel(t)
	cases := []struct {
		name string
		msg  []byte
		want string
	}{
		{"empty", []byte(""), "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"abc", []byte("abc"), "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{"a_64", bytes.Repeat([]byte("a"), 64), "ffe054fe7ae0cb6dc65c3af9b61d5209f439851db43d0ba5997337df154668eb"},
		{"a_1m", bytes.Repeat([]byte("a"), 1000000), "cdc76e5c9914fb9281a1c7e284d73e67f1809a48a497200e046d39ccc7112cd0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Sha256(tc.msg)
			if !bytes.Equal(got, mustHex(t, tc.want)) {
				t.Fatalf("sha256 mismatch: got %x want %s", got, tc.want)
			}
		})
	}
}

func TestSHA256_RandomBlocks(t *testing.T) {
	maybeParallel(t)
	rounds := 5000
	if testing.Short() {
		rounds = 500
	}
	for i := 0; i < rounds; i++ {
		n := 0
		if i > 0 {
			n = (i * 97) % (16 * 1024)
		}
		msg := make([]byte, n)
		if _, err := rand.Read(msg); err != nil {
			t.Fatalf("rand.Read: %v", err)
		}
		got := Sha256(msg)
		want := sha256.Sum256(msg)
		if !bytes.Equal(got, want[:]) {
			t.Fatalf("sha256 mismatch on round %d", i)
		}
	}
}

func TestSHA512_KnownVectors(t *testing.T) {
	maybeParallel(t)
	cases := []struct {
		name string
		msg  []byte
		want string
	}{
		{"empty", []byte(""), "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
		{"abc", []byte("abc"), "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"},
		{"a_128", bytes.Repeat([]byte("a"), 128), "b73d1929aa615934e61a871596b3f3b33359f42b8175602e89f7e06e5f658a243667807ed300314b95cacdd579f3e33abdfbe351909519a846d465c59582f321"},
		{"a_1m", bytes.Repeat([]byte("a"), 1000000), "e718483d0ce769644e2e42c7bc15b4638e1f98b13b2044285632a803afa973ebde0ff244877ea60a4cb0432ce577c31beb009c5c2c49aa2e4eadb217ad8cc09b"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Sha512(tc.msg)
			if !bytes.Equal(got, mustHex(t, tc.want)) {
				t.Fatalf("sha512 mismatch: got %x want %s", got, tc.want)
			}
		})
	}
}

func TestSHA512_RandomBlocks(t *testing.T) {
	maybeParallel(t)
	rounds := 5000
	if testing.Short() {
		rounds = 500
	}
	for i := 0; i < rounds; i++ {
		n := 0
		if i > 0 {
			n = (i * 193) % (16 * 1024)
		}
		msg := make([]byte, n)
		if _, err := rand.Read(msg); err != nil {
			t.Fatalf("rand.Read: %v", err)
		}
		got := Sha512(msg)
		want := sha512.Sum512(msg)
		if !bytes.Equal(got, want[:]) {
			t.Fatalf("sha512 mismatch on round %d", i)
		}
	}
}
