//go:build integration

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func repoRootRNPath(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repo root from %s", wd)
		}
		dir = parent
	}
}

func buildRNPath(t *testing.T, binDir string) string {
	t.Helper()
	name := "rnpath"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out := filepath.Join(binDir, name)
	gocache := filepath.Join(binDir, ".gocache")
	gotmp := filepath.Join(binDir, ".gotmp")
	_ = os.MkdirAll(gocache, 0o755)
	_ = os.MkdirAll(gotmp, 0o755)
	cmd := exec.Command("go", "build", "-o", out, "./cmd/rnpath")
	cmd.Dir = repoRootRNPath(t)
	cmd.Env = append(os.Environ(),
		"GOCACHE="+gocache,
		"GOTMPDIR="+gotmp,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("build rnpath: %v", err)
	}
	return out
}

func writeMinimalReticulumConfigRNPath(t *testing.T, configDir string) {
	t.Helper()
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir configdir: %v", err)
	}
	// Standalone to avoid shared-instance sockets.
	cfg := strings.Join([]string{
		"[reticulum]",
		"enable_transport = False",
		"share_instance = No",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(configDir, "config"), []byte(cfg), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func runRNPath(t *testing.T, ctx context.Context, bin, configDir, workDir string, args ...string) (string, int) {
	t.Helper()
	c := exec.CommandContext(ctx, bin, args...)
	c.Dir = workDir
	home := filepath.Join(configDir, ".home")
	_ = os.MkdirAll(home, 0o755)
	c.Env = append(os.Environ(),
		"HOME="+home,
		"USERPROFILE="+home,
	)
	out, err := c.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	t.Fatalf("run rnpath: %v\n%s", err, string(out))
	return "", -1
}

func skipIfReticulumUnavailableRNPath(t *testing.T, out string, exitCode int) {
	t.Helper()
	if exitCode == 101 || strings.Contains(out, "Could not start Reticulum") ||
		strings.Contains(out, "operation not permitted") {
		t.Skipf("environment does not allow Reticulum startup; skipping rnpath integration test\n%s", out)
	}
}

func TestRNPathIntegration_JSONTableEmpty(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	root := t.TempDir()
	bin := buildRNPath(t, root)
	cfg := filepath.Join(root, "cfg")
	writeMinimalReticulumConfigRNPath(t, cfg)

	out, code := runRNPath(t, ctx, bin, cfg, root, "--config", cfg, "--table", "--json")
	skipIfReticulumUnavailableRNPath(t, out, code)
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	trim := strings.TrimSpace(out)
	if trim != "[]" {
		t.Fatalf("expected [], got %q", trim)
	}
}

func TestRNPathIntegration_DropAnnouncesLocal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	root := t.TempDir()
	bin := buildRNPath(t, root)
	cfg := filepath.Join(root, "cfg")
	writeMinimalReticulumConfigRNPath(t, cfg)

	out, code := runRNPath(t, ctx, bin, cfg, root, "--config", cfg, "--drop-announces")
	skipIfReticulumUnavailableRNPath(t, out, code)
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "Dropping announce queues on all interfaces") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

