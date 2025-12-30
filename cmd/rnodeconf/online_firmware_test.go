package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func httpResponse(status int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func makeTestFirmwareZIP(t *testing.T, version string, fwHash []byte) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)

	write := func(name string, data []byte) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}

	write("extracted_rnode_firmware.version", []byte(version+" "+hex.EncodeToString(fwHash)+"\n"))
	write("extracted_console_image.bin", []byte("console"))
	write("extracted_rnode_firmware.bin", []byte("firmware"))
	write("extracted_rnode_firmware.boot_app0", []byte("boot_app0"))
	write("extracted_rnode_firmware.bootloader", []byte("bootloader"))
	write("extracted_rnode_firmware.partitions", []byte("partitions"))

	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func TestEnsureOnlineFirmwarePrepared_NoCheckRequiresExplicitVersionOrURL(t *testing.T) {
	paths := storagePaths{
		UpdateDir:    t.TempDir(),
		ExtractedDir: t.TempDir(),
	}
	node := &RNode{Model: 0xA1}

	_, err := ensureOnlineFirmwarePrepared(paths, node, "", "", true, false, true)
	if err == nil {
		t.Fatalf("expected error")
	}
	if ee, ok := asExitError(err); !ok || ee.Code != 98 {
		t.Fatalf("expected exit 98, got %v", err)
	}
}

func TestEnsureOnlineFirmwarePrepared_ResolvesLatestAndVerifiesHashAndExtracts(t *testing.T) {
	// Do not run in parallel: overrides global http transport.
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })

	// Prepare filesystem paths.
	base := t.TempDir()
	paths := storagePaths{
		UpdateDir:    filepath.Join(base, "update"),
		ExtractedDir: filepath.Join(base, "extracted"),
	}
	_ = os.MkdirAll(paths.UpdateDir, 0o755)
	_ = os.MkdirAll(paths.ExtractedDir, 0o755)

	node := &RNode{Model: 0xA1, Platform: ROM_PLATFORM_ESP32}
	filename := Models[node.Model].FWFile
	version := "1.2.3"
	fwHash := bytes.Repeat([]byte{0x11}, 32)
	zipBytes := makeTestFirmwareZIP(t, version, fwHash)
	sum := sha256.Sum256(zipBytes)
	zipHashHex := hex.EncodeToString(sum[:])

	releaseLatest := firmwareReleaseInfo{
		filename: {Version: version, Hash: zipHashHex},
	}
	releaseJSON, _ := json.Marshal(releaseLatest)

	fwURL := firmwareUpdateURL + version + "/" + filename

	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.String() {
		case firmwareReleaseInfoURL:
			return httpResponse(200, releaseJSON), nil
		case fwURL:
			return httpResponse(200, zipBytes), nil
		default:
			return httpResponse(404, []byte("not found")), nil
		}
	})

	got, err := ensureOnlineFirmwarePrepared(paths, node, "", "", true, false, false)
	if err != nil {
		t.Fatalf("ensureOnlineFirmwarePrepared: %v", err)
	}
	if got.Version != version {
		t.Fatalf("version got %q want %q", got.Version, version)
	}
	if !got.IsZIP || !got.Extracted {
		t.Fatalf("expected extracted zip, got zip=%v extracted=%v", got.IsZIP, got.Extracted)
	}
	if !fileExists(filepath.Join(paths.ExtractedDir, "extracted_rnode_firmware.version")) {
		t.Fatalf("expected extracted firmware files")
	}
}

func TestEnsureOnlineFirmwarePrepared_HashMismatchExit96(t *testing.T) {
	// Do not run in parallel: overrides global http transport.
	old := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = old })

	base := t.TempDir()
	paths := storagePaths{
		UpdateDir:    filepath.Join(base, "update"),
		ExtractedDir: filepath.Join(base, "extracted"),
	}
	_ = os.MkdirAll(paths.UpdateDir, 0o755)
	_ = os.MkdirAll(paths.ExtractedDir, 0o755)

	node := &RNode{Model: 0xA1, Platform: ROM_PLATFORM_ESP32}
	filename := Models[node.Model].FWFile
	version := "1.2.3"
	fwHash := bytes.Repeat([]byte{0x11}, 32)
	zipBytes := makeTestFirmwareZIP(t, version, fwHash)

	releaseLatest := firmwareReleaseInfo{
		filename: {Version: version, Hash: strings.Repeat("00", 32)},
	}
	releaseJSON, _ := json.Marshal(releaseLatest)

	fwURL := firmwareUpdateURL + version + "/" + filename

	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.String() {
		case firmwareReleaseInfoURL:
			return httpResponse(200, releaseJSON), nil
		case fwURL:
			return httpResponse(200, zipBytes), nil
		default:
			return httpResponse(404, []byte("not found")), nil
		}
	})

	_, err := ensureOnlineFirmwarePrepared(paths, node, "", "", true, false, false)
	if err == nil {
		t.Fatalf("expected error")
	}
	if ee, ok := asExitError(err); !ok || ee.Code != 96 {
		t.Fatalf("expected exit 96, got %v", err)
	}
}
