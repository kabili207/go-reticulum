package interfaces

import (
	"testing"
)

func TestRNodeMulti_parseTruthy(t *testing.T) {
	t.Parallel()

	if parseTruthy("yes", false) != true {
		t.Fatalf("expected yes=true")
	}
	if parseTruthy("0", true) != false {
		t.Fatalf("expected 0=false")
	}
	if parseTruthy("nope", true) != true {
		t.Fatalf("expected default")
	}
}

func TestRNodeMulti_parseFloatSimple(t *testing.T) {
	t.Parallel()

	if v, ok := parseFloatSimple("1.25"); !ok || v != 1.25 {
		t.Fatalf("unexpected parseFloatSimple result: %v %v", v, ok)
	}
	if _, ok := parseFloatSimple("nope"); ok {
		t.Fatalf("expected parse failure")
	}
}

func TestRNodeMulti_parseRNodeMultiConfig_Minimal(t *testing.T) {
	t.Parallel()

	cfg, err := parseRNodeMultiConfig("rm", map[string]string{
		"port": "/dev/ttyUSB0",
		// Minimal subinterface section encoded by reticulum.go: "sub.<name>.<key>"
		"sub.a.vport":           "0",
		"sub.a.frequency":       "868000000",
		"sub.a.bandwidth":       "125000",
		"sub.a.txpower":         "10",
		"sub.a.spreadingfactor": "7",
		"sub.a.codingrate":      "5",
	})
	if err != nil {
		t.Fatalf("parseRNodeMultiConfig: %v", err)
	}
	if cfg.Port != "/dev/ttyUSB0" {
		t.Fatalf("unexpected port %q", cfg.Port)
	}
	if len(cfg.Subinterfaces) != 1 {
		t.Fatalf("expected 1 subinterface, got %d", len(cfg.Subinterfaces))
	}
}
