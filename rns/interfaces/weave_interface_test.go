package interfaces

import (
	"testing"
	"time"
)

func TestWeave_parseWeaveConfig_Minimal(t *testing.T) {
	t.Parallel()

	cfg, err := parseWeaveConfig("w", map[string]string{
		"port": "/dev/ttyUSB0",
	})
	if err != nil {
		t.Fatalf("parseWeaveConfig: %v", err)
	}
	if cfg.port != "/dev/ttyUSB0" {
		t.Fatalf("unexpected port %q", cfg.port)
	}
}

func TestWeaveDevice_TaskStats(t *testing.T) {
	t.Parallel()

	var d WeaveDevice
	d.CaptureTask("t1", 1.0)
	d.CaptureTask("t1", 2.0)
	active := d.ActiveTasks(10 * time.Second)
	if len(active) == 0 {
		t.Fatalf("expected active tasks")
	}
}

