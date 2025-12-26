package main

import "testing"

func TestSpeedStr(t *testing.T) {
	if got := speedStr(999, "bps"); got != "999.00 bps" {
		t.Fatalf("got %q", got)
	}
	if got := speedStr(1000, "bps"); got != "1.00 kbps" {
		t.Fatalf("got %q", got)
	}
	if got := speedStr(8000, "Bps"); got != "1.00 KBps" {
		t.Fatalf("got %q", got)
	}
}

func TestSortInterfaces_MissingKeys(t *testing.T) {
	ifs := []map[string]any{
		{"name": "a", "bitrate": 200},
		{"name": "b"},               // missing bitrate
		{"name": "c", "bitrate": nil}, // nil bitrate
		{"name": "d", "bitrate": 100},
	}
	// Ascending by bitrate should put 0-valued entries first, stable among equals.
	sortInterfaces(ifs, "rate", false)
	if ifs[0]["name"] != "b" || ifs[1]["name"] != "c" || ifs[2]["name"] != "d" || ifs[3]["name"] != "a" {
		t.Fatalf("unexpected order: %v", []any{ifs[0]["name"], ifs[1]["name"], ifs[2]["name"], ifs[3]["name"]})
	}
}

