package rns

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTransportStart_RespondToProbesDoesNotDeadlock(t *testing.T) {
	requireIntegration(t)

	dir, err := os.MkdirTemp("", "rns_transport_start_it_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	// Minimal config that enables probe destination creation during startup, but
	// avoids shared-instance listeners to keep this integration test portable.
	cfg := []byte(
		"[reticulum]\n" +
			"share_instance = no\n" +
			"respond_to_probes = yes\n" +
			"enable_transport = yes\n",
	)
	if err := os.WriteFile(filepath.Join(dir, "config"), cfg, 0o644); err != nil {
		t.Fatalf("WriteFile(config): %v", err)
	}

	// Reticulum is a singleton in-process; restore global state after the test.
	prevInstance := instance
	instance = nil
	t.Cleanup(func() { instance = prevInstance })

	done := make(chan error, 1)
	go func() {
		_, err := NewReticulum(&dir, nil, nil, nil, false, nil)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("NewReticulum: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("NewReticulum did not return; possible startup deadlock when respond_to_probes=yes")
	}
}

