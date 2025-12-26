package interfaces

import (
	"bytes"
	"testing"
)

func TestRNode_KISSEscape(t *testing.T) {
	t.Parallel()

	in := []byte{0x00, KISS_FEND, 0x01, KISS_FESC, 0x02}
	out := kissEscape(in)
	if bytes.Contains(out, []byte{KISS_FEND}) {
		t.Fatalf("unexpected raw FEND in output: %x", out)
	}
}

func TestRNode_BatteryStateString(t *testing.T) {
	t.Parallel()

	r := &RNodeInterface{}
	r.batteryState = 0
	_ = r.BatteryStateString()
}

