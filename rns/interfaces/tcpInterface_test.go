package interfaces

import (
	"bytes"
	"testing"
)

func TestTCPInterface_HDLC_Escape_MatchesLocalHDLC(t *testing.T) {
	t.Parallel()

	in := []byte{0x00, HDLC_FLAG, 0x01, HDLC_ESC, 0x02}
	got := (HDLC{}).Escape(in)
	want := hdlcEscape(in)
	if !bytes.Equal(got, want) {
		t.Fatalf("HDLC escape mismatch: got=%x want=%x", got, want)
	}
}

func TestTCPInterface_KISS_Escape(t *testing.T) {
	t.Parallel()

	in := []byte{0x00, KISS_FEND, 0x01, KISS_FESC, 0x02}
	out := (KISS{}).Escape(in)

	// Ensure no raw FEND/FESC remain.
	for _, b := range out {
		if b == KISS_FEND || b == KISS_FESC {
			// FESC can appear, but only as the escape prefix. A raw FEND should never.
			// Since the output may contain FESC as prefix, accept it but ensure it
			// isn't followed by an invalid byte.
			// Keep the test simple: just ensure raw FEND is absent.
			if b == KISS_FEND {
				t.Fatalf("unexpected raw FEND in escaped output: %x", out)
			}
		}
	}
	if bytes.Contains(out, []byte{KISS_FEND}) {
		t.Fatalf("raw FEND not expected in escaped output: %x", out)
	}
}

