package main

import "testing"

func TestAllSameRune(t *testing.T) {
	if allSameRune("", 'v') {
		t.Fatalf("expected false for empty")
	}
	if !allSameRune("vvv", 'v') {
		t.Fatalf("expected true")
	}
	if allSameRune("vvq", 'v') {
		t.Fatalf("expected false")
	}
}

func TestExpandCountFlags(t *testing.T) {
	got := expandCountFlags([]string{"-vvv", "-qq", "--x", "-vq"})
	want := []string{"-v", "-v", "-v", "-q", "-q", "--x", "-vq"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}

func TestToInt(t *testing.T) {
	if toInt(int64(3)) != 3 {
		t.Fatalf("unexpected")
	}
	if toInt(float64(2.9)) != 2 {
		t.Fatalf("unexpected")
	}
}

func TestToUint64(t *testing.T) {
	if toUint64(int64(-1)) != 0 {
		t.Fatalf("expected clamp to 0")
	}
	if toUint64(float64(3.2)) != 3 {
		t.Fatalf("unexpected")
	}
}

func TestToFloat(t *testing.T) {
	if toFloat(int(3)) != 3 {
		t.Fatalf("unexpected")
	}
	if toFloat(uint64(7)) != 7 {
		t.Fatalf("unexpected")
	}
}

