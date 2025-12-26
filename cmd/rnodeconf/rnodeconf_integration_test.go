//go:build integration

package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func repoRootRNodeconf(t *testing.T) string {
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

var (
	rnodeconfBinOnce sync.Once
	rnodeconfBinPath string
	rnodeconfBinErr  error
)

func getRNodeconfBin(t *testing.T) string {
	t.Helper()
	rnodeconfBinOnce.Do(func() {
		repo := repoRootRNodeconf(t)
		binDir, err := os.MkdirTemp(filepath.Join(repo, ".tmp"), "rnodeconf-bin-")
		if err != nil {
			rnodeconfBinErr = err
			return
		}
		name := "rnodeconf"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		out := filepath.Join(binDir, name)
		gocache := filepath.Join(binDir, ".gocache")
		gotmp := filepath.Join(binDir, ".gotmp")
		_ = os.MkdirAll(gocache, 0o755)
		_ = os.MkdirAll(gotmp, 0o755)
		cmd := exec.Command("go", "build", "-o", out, "./cmd/rnodeconf")
		cmd.Dir = repo
		cmd.Env = append(os.Environ(),
			"GOCACHE="+gocache,
			"GOTMPDIR="+gotmp,
		)
		if err := cmd.Run(); err != nil {
			rnodeconfBinErr = err
			return
		}
		rnodeconfBinPath = out
	})
	if rnodeconfBinErr != nil {
		t.Fatalf("build rnodeconf: %v", rnodeconfBinErr)
	}
	return rnodeconfBinPath
}

func newITestRoot(t *testing.T) string {
	t.Helper()
	repo := repoRootRNodeconf(t)
	root, err := os.MkdirTemp(filepath.Join(repo, ".tmp"), "rnodeconf-it-")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	return root
}

func runRNodeconf(t *testing.T, ctx context.Context, bin, configDir, workDir string, args ...string) (string, int) {
	t.Helper()
	c := exec.CommandContext(ctx, bin, args...)
	c.Dir = workDir

	home := filepath.Join(configDir, ".home")
	rnodeconfDir := filepath.Join(configDir, "cfg")
	_ = os.MkdirAll(home, 0o755)
	_ = os.MkdirAll(rnodeconfDir, 0o755)
	c.Env = append(os.Environ(),
		"HOME="+home,
		"USERPROFILE="+home,
		"RNODECONF_DIR="+rnodeconfDir,
	)
	// Allow commands that prompt for Enter (Python-style UX) to proceed in tests.
	c.Stdin = strings.NewReader("\n")

	out, err := c.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	t.Fatalf("run rnodeconf: %v\n%s", err, string(out))
	return "", -1
}

func TestRNodeconfIntegration_HelpAndVersion(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root, "--help")
	if code != 0 {
		t.Fatalf("help exit=%d\n%s", code, out)
	}
	// Hidden flags should not appear.
	if strings.Contains(out, "get-target-firmware-hash") || strings.Contains(out, "get-firmware-hash") {
		t.Fatalf("expected hidden flags not to appear in help:\n%s", out)
	}
	if !strings.Contains(out, "RNode Configuration and firmware utility") {
		t.Fatalf("unexpected help output:\n%s", out)
	}

	out, code = runRNodeconf(t, ctx, bin, root, root, "--version")
	if code != 0 {
		t.Fatalf("version exit=%d\n%s", code, out)
	}
	if !strings.HasPrefix(out, "rnodeconf ") {
		t.Fatalf("unexpected version output: %q", out)
	}
}

func TestRNodeconfIntegration_UnknownFlagExit2(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root, "--nope")
	if code != 2 {
		t.Fatalf("expected exit 2, got %d\n%s", code, out)
	}
}

func TestRNodeconfIntegration_FlashMissingParamsExit68(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root, "--flash")
	if code != 68 {
		t.Fatalf("expected exit 68, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "Missing parameters") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestRNodeconfIntegration_RNodeconfDirNoAccessExit99(t *testing.T) {
	repo := repoRootRNodeconf(t)
	root, err := os.MkdirTemp(filepath.Join(repo, ".tmp"), "rnodeconf-it-")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	bin := getRNodeconfBin(t)

	// Create a file where RNODECONF_DIR expects a directory.
	bad := filepath.Join(root, "cfg")
	if err := os.WriteFile(bad, []byte("nope"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c := exec.CommandContext(ctx, bin, "--clear-cache")
	c.Dir = root
	home := filepath.Join(root, ".home")
	_ = os.MkdirAll(home, 0o755)
	c.Env = append(os.Environ(),
		"HOME="+home,
		"USERPROFILE="+home,
		"RNODECONF_DIR="+bad,
	)
	outBytes, err := c.CombinedOutput()
	out := string(outBytes)
	code := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			t.Fatalf("run rnodeconf: %v\n%s", err, out)
		}
	}
	if code != 99 {
		t.Fatalf("expected exit 99, got %d\n%s", code, out)
	}
}

func TestRNodeconfIntegration_TrustKeyStoresFile(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	k, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("gen rsa: %v", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&k.PublicKey)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root, "--trust-key", hex.EncodeToString(pubDER))
	if code != 0 {
		t.Fatalf("trust-key exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "Trusting key:") || !strings.Contains(out, "Stored at:") {
		t.Fatalf("unexpected output:\n%s", out)
	}
	// Ensure something was written to trusted_keys.
	tkDir := filepath.Join(root, "cfg", "trusted_keys")
	entries, err := os.ReadDir(tkDir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	found := false
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".pubkey") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a .pubkey file in %s", tkDir)
	}
}

func TestRNodeconfIntegration_TrustKeyInvalidDERExit1(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root, "--trust-key", "00")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "invalid trusted key") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestRNodeconfIntegration_SimPort_InfoAndConfig(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root, "--info", "--port", "sim://esp32")
	if code != 0 {
		t.Fatalf("info exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "Device:") || !strings.Contains(out, "Firmware:") {
		t.Fatalf("unexpected info output:\n%s", out)
	}
	if !strings.Contains(out, "Device signature: Unverified") {
		t.Fatalf("expected unverified signature in output:\n%s", out)
	}

	out, code = runRNodeconf(t, ctx, bin, root, root, "--config", "--port", "sim://esp32")
	if code != 0 {
		t.Fatalf("config exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "Device configuration") && !strings.Contains(out, "WiFi") && !strings.Contains(out, "Config") {
		// output format is verbose and may change; just ensure command didn't fail.
		t.Fatalf("unexpected config output:\n%s", out)
	}
}

func TestRNodeconfIntegration_SimPort_ExtractBranches(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root, "--extract", "--port", "sim://esp32?extract=fail182")
	if code != 182 {
		t.Fatalf("expected exit 182, got %d\n%s", code, out)
	}

	out, code = runRNodeconf(t, ctx, bin, root, root, "--extract", "--port", "sim://esp32?extract=fail180")
	if code != 180 {
		t.Fatalf("expected exit 180, got %d\n%s", code, out)
	}

	out, code = runRNodeconf(t, ctx, bin, root, root, "--extract", "--port", "sim://avr")
	if code != 170 {
		t.Fatalf("expected exit 170, got %d\n%s", code, out)
	}
}

func TestRNodeconfIntegration_SimPort_FlashBranchExit1(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// flash requires extracted firmware presence; create it via simulated extraction first.
	_, code := runRNodeconf(t, ctx, bin, root, root, "--extract", "--port", "sim://esp32")
	if code != 0 {
		t.Fatalf("extract setup failed exit=%d", code)
	}
	out, code := runRNodeconf(t, ctx, bin, root, root, "--flash", "--use-extracted", "--port", "sim://esp32?flash=fail1")
	if code != 1 {
		t.Fatalf("expected exit 1, got %d\n%s", code, out)
	}
}

func TestRNodeconfIntegration_SimPort_EEPROMWipe(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root, "--eeprom-wipe", "--port", "sim://esp32")
	if code != 0 {
		t.Fatalf("eeprom-wipe exit=%d\n%s", code, out)
	}
	if !strings.Contains(out, "EEPROM wipe completed") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestRNodeconfIntegration_SimPort_NotProvisionedExit77_79(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root, "--firmware-hash", strings.Repeat("00", 32), "--port", "sim://esp32?provisioned=0")
	if code != 77 {
		t.Fatalf("expected exit 77, got %d\n%s", code, out)
	}

	out, code = runRNodeconf(t, ctx, bin, root, root, "--sign", "--port", "sim://esp32?provisioned=0")
	if code != 79 {
		t.Fatalf("expected exit 79, got %d\n%s", code, out)
	}
}

func TestRNodeconfIntegration_SimPort_HiddenHashFlagsWork(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root, "-K", "--port", "sim://esp32")
	if code != 0 {
		t.Fatalf("-K exit=%d\n%s", code, out)
	}
	out, code = runRNodeconf(t, ctx, bin, root, root, "-L", "--port", "sim://esp32")
	if code != 0 {
		t.Fatalf("-L exit=%d\n%s", code, out)
	}
}

func TestRNodeconfIntegration_SimPort_UpdateAndAutoinstall(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Prepare extracted firmware.
	_, code := runRNodeconf(t, ctx, bin, root, root, "--extract", "--port", "sim://esp32")
	if code != 0 {
		t.Fatalf("extract setup failed exit=%d", code)
	}

	out, code := runRNodeconf(t, ctx, bin, root, root, "--update", "--use-extracted", "--port", "sim://esp32")
	if code != 0 {
		t.Fatalf("update exit=%d\n%s", code, out)
	}

	// Autoinstall also ROM bootstraps when parameters provided.
	// NOTE: use a fresh sim port instance to avoid cross-test state issues.
	out, code = runRNodeconf(t, ctx, bin, root, root,
		"--autoinstall", "--use-extracted",
		"--product", "0x03", "--model", "A1", "--hwrev", "1",
		"--port", "sim://esp32?session=autoinstall",
	)
	if code != 0 {
		t.Fatalf("autoinstall exit=%d\n%s", code, out)
	}
}

func TestRNodeconfIntegration_SimPort_SettingsAndEEPROMBackup(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, code := runRNodeconf(t, ctx, bin, root, root,
		"--bluetooth-on",
		"--wifi", "OFF",
		"--display", "10",
		"--np", "10",
		"--eeprom-backup",
		"--port", "sim://esp32",
	)
	if code != 0 {
		t.Fatalf("settings exit=%d\n%s", code, out)
	}

	// Ensure EEPROM backup was created in RNODECONF_DIR.
	eepromDir := filepath.Join(root, "cfg", "eeprom")
	entries, err := os.ReadDir(eepromDir)
	if err != nil {
		t.Fatalf("readdir: %v", err)
	}
	found := false
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "eeprom_") && strings.HasSuffix(e.Name(), ".bin") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected EEPROM backup file in %s", eepromDir)
	}
}

func TestRNodeconfIntegration_SimPort_MissingExternalFlashTools(t *testing.T) {
	root := newITestRoot(t)
	bin := getRNodeconfBin(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// AVR: simulate missing avrdude.
	out, code := runRNodeconf(t, ctx, bin, root, root, "--flash", "--fw-url", "file:///tmp/rnode_firmware.hex", "--port", "sim://avr?flash=missing_avrdude")
	if code != 0 {
		t.Fatalf("expected exit 0 (python graceful_exit default), got %d\n%s", code, out)
	}
	if !strings.Contains(out, "avrdude") {
		t.Fatalf("expected avrdude guidance:\n%s", out)
	}

	// NRF52: simulate missing adafruit-nrfutil.
	out, code = runRNodeconf(t, ctx, bin, root, root, "--flash", "--fw-url", "file:///tmp/fw.zip", "--port", "sim://nrf52?flash=missing_nrfutil")
	if code != 0 {
		t.Fatalf("expected exit 0, got %d\n%s", code, out)
	}
	if !strings.Contains(out, "adafruit-nrfutil") {
		t.Fatalf("expected nrfutil guidance:\n%s", out)
	}
}
