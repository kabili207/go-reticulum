package cryptography

import (
	"os"
	"testing"
)

func maybeParallel(t *testing.T) {
	t.Helper()
	if os.Getenv("RNS_INTEGRATION") == "" {
		t.Parallel()
	}
}

