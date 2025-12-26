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
	"testing"
	"time"
)

func repoRootRNID(t *testing.T) string {
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

func buildRNID(t *testing.T, binDir string) string {
	t.Helper()
	name := "rnid"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out := filepath.Join(binDir, name)
	gocache := filepath.Join(binDir, ".gocache")
	gotmp := filepath.Join(binDir, ".gotmp")
	if err := os.MkdirAll(gocache, 0o755); err != nil {
		t.Fatalf("mkdir gocache: %v", err)
	}
	if err := os.MkdirAll(gotmp, 0o755); err != nil {
		t.Fatalf("mkdir gotmp: %v", err)
	}
	cmd := exec.Command("go", "build", "-o", out, "./cmd/rnid")
	cmd.Dir = repoRootRNID(t)
	cmd.Env = append(os.Environ(),
		"GOCACHE="+gocache,
		"GOTMPDIR="+gotmp,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build rnid: %v", err)
	}
	return out
}

func writeMinimalReticulumConfigRNID(t *testing.T, configDir string) {
	t.Helper()
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir configdir: %v", err)
	}
	// Keep it offline and avoid shared instance RPC to reduce sandbox friction.
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

func runRNID(t *testing.T, ctx context.Context, bin, configDir, workDir string, args ...string) (stdout string, exitCode int) {
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
	t.Fatalf("run rnid: %v\n%s", err, string(out))
	return "", -1
}

func skipIfReticulumUnavailable(t *testing.T, out string, exitCode int) {
	t.Helper()
	if exitCode == 101 || strings.Contains(out, "Could not start Reticulum") ||
		strings.Contains(out, "operation not permitted") ||
		strings.Contains(out, "No interfaces could process") {
		t.Skipf("environment does not allow Reticulum startup; skipping rnid integration test\n%s", out)
	}
}

func TestRNIDIntegration_EncryptDecrypt_SignValidate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in -short mode")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	root := t.TempDir()
	bin := buildRNID(t, root)
	cfg := filepath.Join(root, "cfg")
	writeMinimalReticulumConfigRNID(t, cfg)

	work := filepath.Join(root, "work")
	if err := os.MkdirAll(work, 0o755); err != nil {
		t.Fatal(err)
	}

	identityPath := filepath.Join(work, "id")
	out, code := runRNID(t, ctx, bin, cfg, work, "--config", cfg, "--generate", identityPath)
	skipIfReticulumUnavailable(t, out, code)
	if code != 0 {
		t.Fatalf("generate exit=%d\n%s", code, out)
	}
	if _, err := os.Stat(identityPath); err != nil {
		t.Fatalf("expected identity file: %v", err)
	}

	plain := []byte("hello rnid integration\n")
	inPath := filepath.Join(work, "in.txt")
	if err := os.WriteFile(inPath, plain, 0o644); err != nil {
		t.Fatal(err)
	}

	encPath := filepath.Join(work, "in.txt."+encExt)
	out, code = runRNID(t, ctx, bin, cfg, work,
		"--config", cfg,
		"--identity", identityPath,
		"--encrypt", inPath,
		"--write", encPath,
	)
	skipIfReticulumUnavailable(t, out, code)
	if code != 0 {
		t.Fatalf("encrypt exit=%d\n%s", code, out)
	}

	decPath := filepath.Join(work, "out.txt")
	out, code = runRNID(t, ctx, bin, cfg, work,
		"--config", cfg,
		"--identity", identityPath,
		"--decrypt", encPath,
		"--write", decPath,
	)
	skipIfReticulumUnavailable(t, out, code)
	if code != 0 {
		t.Fatalf("decrypt exit=%d\n%s", code, out)
	}
	got, err := os.ReadFile(decPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("decrypt mismatch: got %q want %q", got, plain)
	}

	sigPath := filepath.Join(work, "in.txt."+sigExt)
	out, code = runRNID(t, ctx, bin, cfg, work,
		"--config", cfg,
		"--identity", identityPath,
		"--sign", inPath,
		"--write", sigPath,
	)
	skipIfReticulumUnavailable(t, out, code)
	if code != 0 {
		t.Fatalf("sign exit=%d\n%s", code, out)
	}

	// Validate with explicit read.
	out, code = runRNID(t, ctx, bin, cfg, work,
		"--config", cfg,
		"--identity", identityPath,
		"--validate", sigPath,
		"--read", inPath,
	)
	skipIfReticulumUnavailable(t, out, code)
	if code != 0 {
		t.Fatalf("validate(exit=%d) expected 0\n%s", code, out)
	}

	// Validate without --read (derive from .rsg).
	out, code = runRNID(t, ctx, bin, cfg, work,
		"--config", cfg,
		"--identity", identityPath,
		"--validate", sigPath,
	)
	skipIfReticulumUnavailable(t, out, code)
	if code != 0 {
		t.Fatalf("validate(derive) exit=%d expected 0\n%s", code, out)
	}

	// Tamper input and ensure invalid signature maps to exit 22 (Python parity).
	badPath := filepath.Join(work, "bad.txt")
	if err := os.WriteFile(badPath, []byte("tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, code = runRNID(t, ctx, bin, cfg, work,
		"--config", cfg,
		"--identity", identityPath,
		"--validate", sigPath,
		"--read", badPath,
	)
	skipIfReticulumUnavailable(t, out, code)
	if code != 22 {
		t.Fatalf("validate(tampered) exit=%d expected 22\n%s", code, out)
	}
}

