package main

import (
	"os"
	"testing"
)

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("RNS_INTEGRATION") == "" {
		t.Skip("set RNS_INTEGRATION=1 to run integration tests")
	}
}

func TestIntegration_EchoExample_RequestGetsProof(t *testing.T) {
	requireIntegration(t)
	TestEchoExample_RequestGetsProof(t)
}
