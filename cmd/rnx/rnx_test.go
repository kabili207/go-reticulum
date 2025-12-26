package main

import (
	"reflect"
	"testing"
)

func TestExpandCountFlags(t *testing.T) {
	in := []string{"-vvv", "--quiet", "-qq", "dest", "cmd", "-v", "-q"}
	out := expandCountFlags(in)
	want := []string{"-v", "-v", "-v", "--quiet", "-q", "-q", "dest", "cmd", "-v", "-q"}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("expanded args mismatch:\n got: %#v\nwant: %#v", out, want)
	}
}

func TestSplitCommand_Basic(t *testing.T) {
	args, err := splitCommand(`echo "hello world" 'x y' z\\ w`)
	if err != nil {
		t.Fatalf("splitCommand error: %v", err)
	}
	want := []string{"echo", "hello world", "x y", "z\\", "w"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("args mismatch:\n got: %#v\nwant: %#v", args, want)
	}
}

func TestSplitCommand_Unterminated(t *testing.T) {
	_, err := splitCommand(`echo "nope`)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPrettyTime_NonVerbose_UsesAnd(t *testing.T) {
	// 1 day, 2 hours, 3 minutes, 4 seconds
	got := prettyTime(24*3600+2*3600+3*60+4, false)
	want := "1d, 2h, 3m and 4.0s"
	if got != want {
		t.Fatalf("prettyTime mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestPrettyTime_Verbose_UsesAnd(t *testing.T) {
	got := prettyTime(2*3600+1, true)
	want := "2 hours and 1.0 second"
	if got != want {
		t.Fatalf("prettyTime mismatch:\n got: %q\nwant: %q", got, want)
	}
}

