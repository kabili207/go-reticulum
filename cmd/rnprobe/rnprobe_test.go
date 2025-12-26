package main

import "testing"

func TestConfigDirOrNil(t *testing.T) {
	if configDirOrNil("") != nil {
		t.Fatalf("expected nil")
	}
	if configDirOrNil("   ") != nil {
		t.Fatalf("expected nil")
	}
	p := configDirOrNil("/tmp/x")
	if p == nil || *p != "/tmp/x" {
		t.Fatalf("unexpected %v", p)
	}
}

func TestEffectiveTimeout(t *testing.T) {
	if got := effectiveTimeout(5, 99); got != 5 {
		t.Fatalf("got %v", got)
	}
	if got := effectiveTimeout(0, 2.5); got != DefaultTimeout+2.5 {
		t.Fatalf("got %v", got)
	}
	if got := effectiveTimeout(-1, 1); got != DefaultTimeout+1 {
		t.Fatalf("got %v", got)
	}
}

