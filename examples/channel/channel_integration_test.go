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

func TestIntegration_ChannelExample_StringMessage_RoundTrip(t *testing.T) {
	requireIntegration(t)
	TestChannelExample_StringMessage_RoundTrip(t)
}
