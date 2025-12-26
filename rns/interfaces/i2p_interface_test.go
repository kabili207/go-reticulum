package interfaces

import (
	"bytes"
	"testing"
)

func TestI2P_parseI2PInterfaceConfig_Defaults(t *testing.T) {
	t.Parallel()

	cfg, err := parseI2PInterfaceConfig("i2p", map[string]string{
		"sam_address": "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("parseI2PInterfaceConfig: %v", err)
	}
	if cfg.SAMHost != "127.0.0.1" {
		t.Fatalf("unexpected sam host: %#v", cfg.SAMHost)
	}
}

func TestI2P_KISSEscape_NoRawFEND_FESC(t *testing.T) {
	t.Parallel()

	in := []byte{0x00, i2pKISSFEND, 0x01, i2pKISSFESC, 0x02}
	out := i2pKISSEscape(in)
	if bytes.Contains(out, []byte{i2pKISSFEND}) {
		t.Fatalf("unexpected raw FEND in output: %x", out)
	}
}
