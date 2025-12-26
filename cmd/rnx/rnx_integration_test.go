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

func repoRootRNX(t *testing.T) string {
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

func buildRNX(t *testing.T, binDir string) string {
	t.Helper()
	name := "rnx"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out := filepath.Join(binDir, name)
	gocache := filepath.Join(binDir, ".gocache")
	gotmp := filepath.Join(binDir, ".gotmp")
	_ = os.MkdirAll(gocache, 0o755)
	_ = os.MkdirAll(gotmp, 0o755)
	cmd := exec.Command("go", "build", "-o", out, "./cmd/rnx")
	cmd.Dir = repoRootRNX(t)
	cmd.Env = append(os.Environ(),
		"GOCACHE="+gocache,
		"GOTMPDIR="+gotmp,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("build rnx: %v", err)
	}
	return out
}

func runRNX(t *testing.T, ctx context.Context, bin, configDir, workDir string, args ...string) (string, int) {
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
	t.Fatalf("run rnx: %v\n%s", err, string(out))
	return "", -1
}

func TestRNXIntegration_HelpAndVersion(t *testing.T) {
	root := t.TempDir()
	bin := buildRNX(t, root)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, code := runRNX(t, ctx, bin, root, root, "--help")
	if code != 0 {
		t.Fatalf("help exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "Reticulum Remote Execution Utility") {
		t.Fatalf("unexpected help output:\n%s", out)
	}

	out, code = runRNX(t, ctx, bin, root, root, "--version")
	if code != 0 {
		t.Fatalf("version exit=%d\n%s", code, out)
	}
	if !strings.HasPrefix(out, "rnx ") {
		t.Fatalf("unexpected version output: %q", out)
	}
}

func TestRNXIntegration_UnknownFlagExit2(t *testing.T) {
	root := t.TempDir()
	bin := buildRNX(t, root)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, code := runRNX(t, ctx, bin, root, root, "--nope")
	if code != 2 {
		t.Fatalf("expected exit 2, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "unrecognized arguments") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

