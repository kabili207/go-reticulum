package main

import (
	"os"
	"testing"
)

func TestIntegration_Minimal_ProgramSetup(t *testing.T) {
	if os.Getenv("RNS_INTEGRATION") == "" {
		t.Skip("set RNS_INTEGRATION=1 to run integration tests")
	}
	TestProgramSetupCreatesDestination(t)
}
