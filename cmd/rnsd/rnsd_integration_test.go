//go:build integration

package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

func repoRootRNSD(t *testing.T) string {
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

func buildRNSD(t *testing.T, binDir string) string {
	t.Helper()
	name := "rnsd"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out := filepath.Join(binDir, name)
	gocache := filepath.Join(binDir, ".gocache")
	gotmp := filepath.Join(binDir, ".gotmp")
	_ = os.MkdirAll(gocache, 0o755)
	_ = os.MkdirAll(gotmp, 0o755)
	cmd := exec.Command("go", "build", "-o", out, "./cmd/rnsd")
	cmd.Dir = repoRootRNSD(t)
	cmd.Env = append(os.Environ(),
		"GOCACHE="+gocache,
		"GOTMPDIR="+gotmp,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("build rnsd: %v", err)
	}
	return out
}

func writeMinimalReticulumConfigRNSD(t *testing.T, configDir string) {
	t.Helper()
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir configdir: %v", err)
	}
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

func runRNSD(t *testing.T, ctx context.Context, bin, configDir, workDir string, args ...string) (string, int) {
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
	t.Fatalf("run rnsd: %v\n%s", err, string(out))
	return "", -1
}

func skipIfReticulumUnavailableRNSD(t *testing.T, out string, exitCode int) {
	t.Helper()
	if exitCode == 1 && (strings.Contains(out, "Error starting rnsd") || strings.Contains(out, "operation not permitted")) {
		t.Skipf("environment does not allow Reticulum startup; skipping rnsd integration test\n%s", out)
	}
}

func TestRNSDIntegration_ExampleConfig(t *testing.T) {
	root := t.TempDir()
	bin := buildRNSD(t, root)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, code := runRNSD(t, ctx, bin, root, root, "--exampleconfig")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if out != exampleRNSConfig {
		t.Fatalf("unexpected example config output")
	}
}

func TestRNSDIntegration_Version(t *testing.T) {
	root := t.TempDir()
	bin := buildRNSD(t, root)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, code := runRNSD(t, ctx, bin, root, root, "--version")
	if code != 0 {
		t.Fatalf("exit=%d\n%s", code, out)
	}
	if !strings.HasPrefix(out, "rnsd ") {
		t.Fatalf("unexpected version output: %q", out)
	}
}

func TestRNSDIntegration_ServiceCreatesLogfile(t *testing.T) {
	root := t.TempDir()
	bin := buildRNSD(t, root)
	cfg := filepath.Join(root, "cfg")
	writeMinimalReticulumConfigRNSD(t, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c := exec.CommandContext(ctx, bin, "--config", cfg, "--service")
	c.Dir = root
	home := filepath.Join(cfg, ".home")
	_ = os.MkdirAll(home, 0o755)
	c.Env = append(os.Environ(),
		"HOME="+home,
		"USERPROFILE="+home,
	)
	var out bytes.Buffer
	c.Stdout = &out
	c.Stderr = &out
	if err := c.Start(); err != nil {
		t.Fatalf("start rnsd: %v", err)
	}
	t.Cleanup(func() {
		_ = c.Process.Signal(syscall.SIGTERM)
		_ = c.Wait()
	})

	logPath := filepath.Join(cfg, "logfile")
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(logPath); err == nil {
			goto found
		}
		time.Sleep(100 * time.Millisecond)
	}
	// If it exited quickly, report startup failure.
	if c.ProcessState != nil && c.ProcessState.Exited() {
		code := c.ProcessState.ExitCode()
		skipIfReticulumUnavailableRNSD(t, out.String(), code)
		t.Fatalf("rnsd service exited unexpectedly (exit=%d)\n%s", code, out.String())
	}
	t.Fatalf("expected logfile to exist; output:\n%s", out.String())

found:
	_ = c.Process.Signal(syscall.SIGTERM)
	_ = c.Wait()
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("expected logfile to exist: %v", err)
	}
}
