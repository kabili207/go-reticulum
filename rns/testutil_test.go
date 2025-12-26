package rns

import (
	"os"
	"testing"
)

// When running integration tests, this package uses global mutable state (Transport,
// destinations, links, caches). Running unit tests in parallel can make the suite
// flaky. Disable t.Parallel() while RNS_INTEGRATION=1.
func maybeParallel(t *testing.T) {
	t.Helper()
	if os.Getenv("RNS_INTEGRATION") == "" {
		t.Parallel()
	}
}

