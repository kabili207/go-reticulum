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

func TestNumField_UnsignedCounters(t *testing.T) {
	m := map[string]any{
		"u64": uint64(123),
		"u32": uint32(456),
		"u":   uint(789),
	}
	if got, ok := numField(m, "u64"); !ok || got != 123 {
		t.Fatalf("u64: got=%d ok=%v", got, ok)
	}
	if got, ok := numField(m, "u32"); !ok || got != 456 {
		t.Fatalf("u32: got=%d ok=%v", got, ok)
	}
	if got, ok := numField(m, "u"); !ok || got != 789 {
		t.Fatalf("u: got=%d ok=%v", got, ok)
	}
}

func TestSortInterfaces_MissingKeys(t *testing.T) {
	ifs := []map[string]any{
		{"name": "a", "bitrate": 200},
		{"name": "b"},                 // missing bitrate
		{"name": "c", "bitrate": nil}, // nil bitrate
		{"name": "d", "bitrate": 100},
	}
	// Ascending by bitrate should put 0-valued entries first, stable among equals.
	sortInterfaces(ifs, "rate", false)
	if ifs[0]["name"] != "b" || ifs[1]["name"] != "c" || ifs[2]["name"] != "d" || ifs[3]["name"] != "a" {
		t.Fatalf("unexpected order: %v", []any{ifs[0]["name"], ifs[1]["name"], ifs[2]["name"], ifs[3]["name"]})
	}
}
