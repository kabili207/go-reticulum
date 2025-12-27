package main

import (
	"os"
	"testing"
)

func TestIntegration_InterfaceExample_SnippetAndModes(t *testing.T) {
	if os.Getenv("RNS_INTEGRATION") == "" {
		t.Skip("set RNS_INTEGRATION=1 to run integration tests")
	}
	TestConfigSnippetContainsDefaults(t)
	TestParseMode(t)
}
