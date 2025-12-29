package main

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rns "github.com/svanichkin/go-reticulum/rns"
)

func TestExpandCountFlags(t *testing.T) {
	t.Run("expands -vvv", func(t *testing.T) {
		got := expandCountFlags([]string{"-vvv", "file", "dest"})
		want := []string{"-v", "-v", "-v", "file", "dest"}
		if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
			t.Fatalf("got %q want %q", got, want)
		}
	})

	t.Run("expands -qq", func(t *testing.T) {
		got := expandCountFlags([]string{"-qq", "--fetch"})
		want := []string{"-q", "-q", "--fetch"}
		if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
			t.Fatalf("got %q want %q", got, want)
		}
	})

	t.Run("does not touch mixed short flags", func(t *testing.T) {
		got := expandCountFlags([]string{"-vq", "x"})
		want := []string{"-vq", "x"}
		if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
			t.Fatalf("got %q want %q", got, want)
		}
	})

	t.Run("does not touch long flags", func(t *testing.T) {
		got := expandCountFlags([]string{"--vvv", "x"})
		want := []string{"--vvv", "x"}
		if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}

func TestAllSameRune(t *testing.T) {
	if allSameRune("", 'v') {
		t.Fatalf("empty string should be false")
	}
	if !allSameRune("vvv", 'v') {
		t.Fatalf("expected true")
	}
	if allSameRune("vvq", 'v') {
		t.Fatalf("expected false")
	}
}

func TestSizeStr(t *testing.T) {
	// Mirrors python/RNS/Utilities/rncp.py:size_str() formatting.
	if got, want := sizeStr(999, 'B'), "999 B"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got, want := sizeStr(1000, 'B'), "1.00 KB"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got, want := sizeStr(1000, 'b'), "8.00 Kb"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestParseDest(t *testing.T) {
	destLen := (rns.ReticulumTruncatedHashLength / 8) * 2
	good := strings.Repeat("a", destLen)
	b, err := parseDest(good)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b) != destLen/2 {
		t.Fatalf("got %d bytes want %d", len(b), destLen/2)
	}

	_, err = parseDest(good + "aa")
	if err == nil {
		t.Fatalf("expected error for invalid length")
	}
}

func TestExpandPath_Tilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if got, want := expandPath("~"), home; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got, want := expandPath("~/a/b"), filepath.Join(home, "a", "b"); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestLoadAllowedIdentities_FromConfigFilePrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	destLen := (rns.ReticulumTruncatedHashLength / 8) * 2
	cfgHash := strings.Repeat("a", destLen)
	rncpHash := strings.Repeat("b", destLen)

	cfgDir := filepath.Join(home, ".config", "rncp")
	rncpDir := filepath.Join(home, ".rncp")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.MkdirAll(rncpDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "allowed_identities"), []byte(cfgHash+"\nignored\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.WriteFile(filepath.Join(rncpDir, "allowed_identities"), []byte(rncpHash+"\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	allowedIdentityHashes = nil
	if err := loadAllowedIdentities(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(allowedIdentityHashes) != 1 {
		t.Fatalf("got %d allowed identities, want 1", len(allowedIdentityHashes))
	}
	wantBytes, _ := hex.DecodeString(cfgHash)
	if got := hex.EncodeToString(allowedIdentityHashes[0]); got != hex.EncodeToString(wantBytes) {
		t.Fatalf("got %q want %q", got, cfgHash)
	}
}
