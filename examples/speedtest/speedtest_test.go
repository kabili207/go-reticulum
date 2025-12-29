package main

import (
	"testing"

	rns "github.com/svanichkin/go-reticulum/rns"
)

func TestParseTruncatedHashHexLength(t *testing.T) {
	wantLen := (rns.TRUNCATED_HASHLENGTH / 8) * 2
	_, err := parseTruncatedHashHex("00")
	if err == nil {
		t.Fatalf("expected error for too-short hash")
	}
	_, err = parseTruncatedHashHex(makeString('a', wantLen))
	if err != nil {
		t.Fatalf("expected valid hash length, got %v", err)
	}
}

func makeString(ch byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ch
	}
	return string(b)
}

func TestSizeStrUnits(t *testing.T) {
	if got := sizeStr(1024); got == "" {
		t.Fatalf("expected non-empty")
	}
	if got := sizeStr(1, "b"); got == "" {
		t.Fatalf("expected non-empty")
	}
}
