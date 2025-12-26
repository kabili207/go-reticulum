package main

import (
	"strings"
	"testing"
)

func TestExpandCountFlags(t *testing.T) {
	t.Run("expands -vvv", func(t *testing.T) {
		got := expandCountFlags([]string{"-vvv", "x"})
		want := []string{"-v", "-v", "-v", "x"}
		if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
			t.Fatalf("got %q want %q", got, want)
		}
	})

	t.Run("expands -qq", func(t *testing.T) {
		got := expandCountFlags([]string{"-qq", "--exampleconfig"})
		want := []string{"-q", "-q", "--exampleconfig"}
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

