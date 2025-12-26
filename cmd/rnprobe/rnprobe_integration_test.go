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

func repoRootRNProbe(t *testing.T) string {
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

func buildRNProbe(t *testing.T, binDir string) string {
	t.Helper()
	name := "rnprobe"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out := filepath.Join(binDir, name)
	gocache := filepath.Join(binDir, ".gocache")
	gotmp := filepath.Join(binDir, ".gotmp")
	_ = os.MkdirAll(gocache, 0o755)
	_ = os.MkdirAll(gotmp, 0o755)
	cmd := exec.Command("go", "build", "-o", out, "./cmd/rnprobe")
	cmd.Dir = repoRootRNProbe(t)
	cmd.Env = append(os.Environ(),
		"GOCACHE="+gocache,
		"GOTMPDIR="+gotmp,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("build rnprobe: %v", err)
	}
	return out
}

func writeMinimalReticulumConfigRNProbe(t *testing.T, configDir string) {
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

func runRNProbe(t *testing.T, ctx context.Context, bin, configDir, workDir string, args ...string) (string, int) {
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
	t.Fatalf("run rnprobe: %v\n%s", err, string(out))
	return "", -1
}

func TestRNProbeIntegration_InvalidArgsExit0(t *testing.T) {
	root := t.TempDir()
	bin := buildRNProbe(t, root)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cfg := filepath.Join(root, "cfg")
	writeMinimalReticulumConfigRNProbe(t, cfg)

	// Invalid destination length should be exit code 0 (Python exit()).
	out, code := runRNProbe(t, ctx, bin, cfg, root, "--config", cfg, "app.aspect", "aa")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "Destination length is invalid") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}
