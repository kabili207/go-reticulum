package main

import (
	"encoding/base64"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandCountFlags(t *testing.T) {
	got := expandCountFlags([]string{"-vvv", "-qq", "--x"})
	want := []string{"-v", "-v", "-v", "-q", "-q", "--x"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("got %q want %q", got, want)
	}
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

func TestExpandUser_Tilde(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if got, want := expandUser("~"), home; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got, want := expandUser("~/a/b"), filepath.Join(home, "a", "b"); got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if got := expandUser("~user/x"); got != "~user/x" {
		t.Fatalf("~user expansion should be untouched, got %q", got)
	}
}

func TestDecodeIdentityBase64(t *testing.T) {
	src := []byte("hello world")

	// urlsafe_b64encode with padding
	padded := base64.URLEncoding.EncodeToString(src)
	got, err := decodeIdentityBase64(padded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(src) {
		t.Fatalf("got %q want %q", got, src)
	}

	// urlsafe_b64encode without padding (common CLI form)
	unpadded := strings.TrimRight(padded, "=")
	got, err = decodeIdentityBase64(unpadded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(src) {
		t.Fatalf("got %q want %q", got, src)
	}
}

func TestSpin(t *testing.T) {
	t.Run("immediate success", func(t *testing.T) {
		ok := spin(func() bool { return true }, "msg", 0)
		if !ok {
			t.Fatalf("expected ok")
		}
	})
}

func TestValidateDeriveReadMatchesPythonReplace(t *testing.T) {
	read := ""
	validate := "dir/file.rsg"
	if strings.HasSuffix(strings.ToLower(validate), "."+sigExt) {
		read = strings.Replace(validate, "."+sigExt, "", -1)
	}
	if read != "dir/file" {
		t.Fatalf("got %q", read)
	}
}
