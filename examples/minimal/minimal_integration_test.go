package main

import (
	"os"
	"testing"
)

func TestIntegration_Minimal_ProgramSetup(t *testing.T) {
	if os.Getenv("RNS_INTEGRATION") == "" {
		t.Skip("set RNS_INTEGRATION=1 to run integration tests")
	}
	t.Skip("covered by TestProgramSetupCreatesDestination; integration wrapper disabled to avoid Reticulum singleton conflicts")
}
