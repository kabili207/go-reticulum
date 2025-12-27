package main

import (
	"strings"
	"testing"
)

func TestConfigSnippetContainsDefaults(t *testing.T) {
	s := configSnippet("Example Custom Interface")
	for _, want := range []string{
		"type = ExampleInterface",
		"mode = gateway",
		"port = /dev/ttyUSB0",
		"speed = 115200",
		"databits = 8",
		"parity = none",
		"stopbits = 1",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("snippet missing %q:\n%s", want, s)
		}
	}
}

func TestParseMode(t *testing.T) {
	if got := parseMode("gateway"); got == 0 {
		t.Fatalf("expected non-zero mode for gateway")
	}
	if got := parseMode("unknown"); got == 0 {
		t.Fatalf("expected fallback mode to be non-zero")
	}
}
