package main

import (
	"os"
	"path/filepath"
	"testing"

	rns "main/rns"
	umsgpack "main/rns/vendor"
)

func TestListFilesFiltersAndSorts(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".hidden"), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	files, err := listFiles(dir)
	if err != nil {
		t.Fatalf("listFiles: %v", err)
	}
	if len(files) != 2 || files[0] != "a.txt" || files[1] != "b.txt" {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestPackFileListChunksRespectsLimit(t *testing.T) {
	files := []string{"one.txt", "two.txt", "three.txt", "four.txt"}

	// Pick a payload limit that should force chunking.
	limit := 16
	chunks, err := packFileListChunks(files, limit)
	if err != nil {
		t.Fatalf("packFileListChunks: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected chunking, got %d chunk(s)", len(chunks))
	}

	var roundtrip []string
	for _, c := range chunks {
		if len(c) > limit {
			t.Fatalf("chunk exceeds limit: %d > %d", len(c), limit)
		}
		var part []string
		if err := umsgpack.Unpackb(c, &part); err != nil {
			t.Fatalf("unpack chunk: %v", err)
		}
		roundtrip = append(roundtrip, part...)
	}

	if len(roundtrip) != len(files) {
		t.Fatalf("roundtrip len mismatch: got %d want %d", len(roundtrip), len(files))
	}
}

func TestFileExistsInDirRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ok.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if !fileExistsInDir(dir, "ok.txt") {
		t.Fatalf("expected ok.txt to exist")
	}
	for _, bad := range []string{"../ok.txt", "..", "/etc/passwd", "sub/ok.txt"} {
		if fileExistsInDir(dir, bad) {
			t.Fatalf("expected %q to be rejected", bad)
		}
	}
}

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
