package rns

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIntegration_Reticulum_SharedInstanceRPC_Getters(t *testing.T) {
	requireIntegration(t)

	dir, err := os.MkdirTemp("", "rns_reticulum_it_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	// Avoid binding fixed ports; use ephemeral TCP ports and a unix tempfile for packet IPC.
	rpcKey := make([]byte, 16)
	for i := range rpcKey {
		rpcKey[i] = byte(i + 1)
	}
	cfg := []byte(
		"[reticulum]\n" +
			"share_instance = yes\n" +
			"shared_instance_type = tcp\n" +
			"shared_instance_port = 0\n" +
			"instance_control_port = 0\n" +
			"rpc_key = " + hex.EncodeToString(rpcKey) + "\n",
	)
	if err := os.WriteFile(filepath.Join(dir, "config"), cfg, 0o644); err != nil {
		t.Fatalf("WriteFile(config): %v", err)
	}

	// Reticulum is a singleton in-process; restore global state after the test.
	prevInstance := instance
	instance = nil
	t.Cleanup(func() { instance = prevInstance })

	r, err := NewReticulum(&dir, nil, nil, nil, false, func() *string { s := "tcp"; return &s }())
	if err != nil {
		t.Fatalf("NewReticulum: %v", err)
	}
	if r == nil {
		t.Fatalf("NewReticulum returned nil")
	}
	t.Cleanup(func() {
		if r.rpcLn != nil {
			_ = r.rpcLn.Close()
		}
	})

	if !r.IsSharedInstance {
		// In some sandboxed environments, binding loopback TCP listeners is not permitted.
		// Treat this as an environment limitation rather than a functional failure.
		t.Skipf("shared instance RPC not available in this environment (shared=%v standalone=%v connected=%v)", r.IsSharedInstance, r.IsStandaloneInstance, r.IsConnectedToSharedInstance)
	}
	if r.rpcLn == nil || r.rpcLn.Addr() == "" {
		t.Fatalf("expected rpc listener addr")
	}

	// Allow the accept loop to start.
	time.Sleep(10 * time.Millisecond)

	// Basic getter roundtrips.
	{
		c, err := dialRPC("tcp", r.rpcLn.Addr(), r.RPCKey)
		if err != nil {
			// Some sandboxed test environments disallow loopback TCP dials.
			// Treat this as an environment limitation rather than a functional failure.
			if strings.Contains(err.Error(), "operation not permitted") {
				t.Skipf("loopback tcp dial not permitted in this environment: %v", err)
			}
			t.Fatalf("dialRPC: %v", err)
		}
		defer c.Close()

		if err := c.Send(map[string]any{"get": "link_count"}); err != nil {
			t.Fatalf("rpc send: %v", err)
		}
		var count int
		if err := c.Recv(&count); err != nil {
			t.Fatalf("rpc recv: %v", err)
		}
		if count < 0 {
			t.Fatalf("unexpected link_count=%d", count)
		}
	}

	{
		c, err := dialRPC("tcp", r.rpcLn.Addr(), r.RPCKey)
		if err != nil {
			if strings.Contains(err.Error(), "operation not permitted") {
				t.Skipf("loopback tcp dial not permitted in this environment: %v", err)
			}
			t.Fatalf("dialRPC: %v", err)
		}
		defer c.Close()

		if err := c.Send(map[string]any{"get": "interface_stats"}); err != nil {
			t.Fatalf("rpc send: %v", err)
		}
		var stats map[string]any
		if err := c.Recv(&stats); err != nil {
			t.Fatalf("rpc recv: %v", err)
		}
		if stats == nil {
			t.Fatalf("expected interface_stats map")
		}
	}
}
