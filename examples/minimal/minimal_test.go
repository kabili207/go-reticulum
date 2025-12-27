package main

import (
	"bytes"
	"testing"

	rns "main/rns"
)

func TestProgramSetupCreatesDestination(t *testing.T) {
	prev := rns.Destinations
	t.Cleanup(func() { rns.Destinations = prev })
	rns.Destinations = nil

	if err := programSetup(nil, bytes.NewBufferString("")); err != nil {
		t.Fatalf("programSetup: %v", err)
	}
	if len(rns.Destinations) == 0 {
		t.Fatalf("expected destinations to be created")
	}
}
