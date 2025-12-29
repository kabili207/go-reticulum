package rns

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReticulumCleanCaches_RemovesStaleFiles(t *testing.T) {
	// Modifies filesystem and uses timestamps; keep serial.

	dir, err := os.MkdirTemp("", "rns_reticulum_cache_*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	r := &Reticulum{
		ResourcePath: filepath.Join(dir, "storage/resources"),
		CachePath:    filepath.Join(dir, "storage/cache"),
	}
	if err := os.MkdirAll(r.ResourcePath, 0o755); err != nil {
		t.Fatalf("MkdirAll(ResourcePath): %v", err)
	}
	if err := os.MkdirAll(r.CachePath, 0o755); err != nil {
		t.Fatalf("MkdirAll(CachePath): %v", err)
	}

	fullHashName := strings.Repeat("a", (IdentityHashLength/8)*2)
	resOld := filepath.Join(r.ResourcePath, fullHashName)
	pktOld := filepath.Join(r.CachePath, fullHashName)
	resNew := filepath.Join(r.ResourcePath, strings.Repeat("b", (IdentityHashLength/8)*2))
	pktNew := filepath.Join(r.CachePath, strings.Repeat("c", (IdentityHashLength/8)*2))

	for _, p := range []string{resOld, pktOld, resNew, pktNew} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", p, err)
		}
	}

	oldTime := time.Now().Add(-DestinationTimeout - 24*time.Hour)
	if err := os.Chtimes(resOld, oldTime, oldTime); err != nil {
		t.Fatalf("Chtimes(resOld): %v", err)
	}
	if err := os.Chtimes(pktOld, oldTime, oldTime); err != nil {
		t.Fatalf("Chtimes(pktOld): %v", err)
	}

	r.cleanCaches()

	if _, err := os.Stat(resOld); !os.IsNotExist(err) {
		t.Fatalf("expected resOld removed, err=%v", err)
	}
	if _, err := os.Stat(pktOld); !os.IsNotExist(err) {
		t.Fatalf("expected pktOld removed, err=%v", err)
	}
	if _, err := os.Stat(resNew); err != nil {
		t.Fatalf("expected resNew present, err=%v", err)
	}
	if _, err := os.Stat(pktNew); err != nil {
		t.Fatalf("expected pktNew present, err=%v", err)
	}
}

