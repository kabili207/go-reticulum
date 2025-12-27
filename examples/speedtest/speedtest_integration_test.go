package main

import (
	"os"
	"testing"
)

func TestIntegration_Speedtest_ParseTruncatedHashHex(t *testing.T) {
	if os.Getenv("RNS_INTEGRATION") == "" {
		t.Skip("set RNS_INTEGRATION=1 to run integration tests")
	}
	TestParseTruncatedHashHexLength(t)
}
