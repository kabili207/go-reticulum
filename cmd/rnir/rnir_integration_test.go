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

func repoRootRNIR(t *testing.T) string {
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

func buildRNIR(t *testing.T, binDir string) string {
	t.Helper()
	name := "rnir"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out := filepath.Join(binDir, name)
	gocache := filepath.Join(binDir, ".gocache")
	gotmp := filepath.Join(binDir, ".gotmp")
	_ = os.MkdirAll(gocache, 0o755)
	_ = os.MkdirAll(gotmp, 0o755)
	cmd := exec.Command("go", "build", "-o", out, "./cmd/rnir")
	cmd.Dir = repoRootRNIR(t)
	cmd.Env = append(os.Environ(),
		"GOCACHE="+gocache,
		"GOTMPDIR="+gotmp,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("build rnir: %v", err)
	}
	return out
}

func writeMinimalReticulumConfigRNIR(t *testing.T, configDir string) {
	t.Helper()
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir configdir: %v", err)
	}
	// Prefer standalone instance to avoid shared-instance socket/network complexity.
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

func runRNIR(t *testing.T, ctx context.Context, bin, configDir, workDir string, args ...string) (string, int) {
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
	t.Fatalf("run rnir: %v\n%s", err, string(out))
	return "", -1
}

func skipIfReticulumUnavailableRNIR(t *testing.T, out string, exitCode int) {
	t.Helper()
	if exitCode == 1 && (strings.Contains(out, "Could not start Reticulum") || strings.Contains(out, "operation not permitted")) {
		t.Skipf("environment does not allow Reticulum startup; skipping rnir integration test\n%s", out)
	}
}

func TestRNIRIntegration_ExampleConfig(t *testing.T) {
	root := t.TempDir()
	bin := buildRNIR(t, root)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, code := runRNIR(t, ctx, bin, root, root, "--exampleconfig")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if out != exampleConfig {
		t.Fatalf("unexpected output:\n%q", out)
	}
}

func TestRNIRIntegration_ServiceModeCreatesLogfile(t *testing.T) {
	root := t.TempDir()
	bin := buildRNIR(t, root)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	cfg := filepath.Join(root, "cfg")
	writeMinimalReticulumConfigRNIR(t, cfg)

	out, code := runRNIR(t, ctx, bin, cfg, root, "--config", cfg, "--service")
	skipIfReticulumUnavailableRNIR(t, out, code)
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if _, err := os.Stat(filepath.Join(cfg, "logfile")); err != nil {
		t.Fatalf("expected logfile to exist: %v", err)
	}
}
