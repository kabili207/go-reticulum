package main

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestRNode_VerifyDeviceSignature_LocalKey(t *testing.T) {
	dir := t.TempDir()
	paths := storagePaths{
		SignerKeyPath:  filepath.Join(dir, "signing.key"),
		TrustedKeysDir: filepath.Join(dir, "trusted_keys"),
	}
	_ = os.MkdirAll(paths.TrustedKeysDir, 0o755)

	// Build a valid provisioned EEPROM chunk.
	product := byte(0x03)
	model := byte(0xA1)
	hwrev := byte(0x01)
	serialNo := []byte{0, 0, 0, 1}
	made := []byte{0, 0, 0, 2}
	info := append([]byte{product, model, hwrev}, append(serialNo, made...)...)
	sum := md5.Sum(info)
	checksum := sum[:]

	// Generate a local 1024-bit signer and sign checksum using PSS+SHA256 (Python parity).
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("gen rsa: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	if err := os.WriteFile(paths.SignerKeyPath, der, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
	checksumHash := sha256.Sum256(checksum)
	sig, err := rsa.SignPSS(rand.Reader, priv, crypto.SHA256, checksumHash[:], nil)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if len(sig) != 128 {
		t.Fatalf("sig len=%d", len(sig))
	}

	eeprom := make([]byte, romAddrSignature+128)
	eeprom[romAddrProduct] = product
	eeprom[romAddrModel] = model
	eeprom[romAddrHWRev] = hwrev
	copy(eeprom[romAddrSerial:romAddrSerial+4], serialNo)
	copy(eeprom[romAddrMade:romAddrMade+4], made)
	copy(eeprom[romAddrChecksum:romAddrChecksum+16], checksum)
	copy(eeprom[romAddrSignature:romAddrSignature+128], sig)
	eeprom = append(eeprom, make([]byte, int(romAddrInfoLock)-len(eeprom)+1)...)
	eeprom[romAddrInfoLock] = romInfoLockByte

	n := &RNode{EEPROM: eeprom}
	n.parseEEPROM()
	if !n.Provisioned {
		t.Fatalf("expected provisioned")
	}
	n.verifyDeviceSignature(paths)
	if !n.SignatureValid || !n.LocallySigned || n.Vendor != "LOCAL" {
		t.Fatalf("unexpected signature state: valid=%v local=%v vendor=%q", n.SignatureValid, n.LocallySigned, n.Vendor)
	}
}

func TestFirmwareCacheReadWrite(t *testing.T) {
	base := t.TempDir()
	paths := storagePaths{UpdateDir: base}
	writeCachedFirmwareVersion(paths, "1.2.3", "fw.zip", hex.EncodeToString(make([]byte, 32)))
	if _, ok := readCachedFirmwareHash(paths, "1.2.3", "fw.zip"); !ok {
		t.Fatalf("expected cached hash")
	}
}

func TestRunExtraction_NonESP32Exit170(t *testing.T) {
	paths := storagePaths{ExtractedDir: t.TempDir()}
	err := runExtraction(&RNode{Detected: true, Platform: ROM_PLATFORM_AVR}, "ttyS0", 115200, paths)
	ee, ok := asExitError(err)
	if !ok || ee.Code != 170 {
		t.Fatalf("expected exitError 170, got (%v,%v)", err, ok)
	}
}

func TestEnsureDeviceSignerKey_CreateAndLoad(t *testing.T) {
	paths := storagePaths{DeviceKeyPath: filepath.Join(t.TempDir(), "device.key")}
	id, err := ensureDeviceSignerKey(paths)
	if err != nil || id == nil {
		t.Fatalf("create: (%v,%v)", id, err)
	}
	id2, err := ensureDeviceSignerKey(paths)
	if err != nil || id2 == nil {
		t.Fatalf("load: (%v,%v)", id2, err)
	}
}

func TestEnsureDeviceSignerKey_LoadErrorExit82(t *testing.T) {
	base := t.TempDir()
	paths := storagePaths{DeviceKeyPath: filepath.Join(base, "device.key")}
	if err := os.WriteFile(paths.DeviceKeyPath, []byte("not an identity"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	err := func() error {
		_, err := ensureDeviceSignerKey(paths)
		return err
	}()
	ee, ok := asExitError(err)
	if !ok || ee.Code != 82 {
		t.Fatalf("expected exitError 82, got (%v,%v)", err, ok)
	}
}
