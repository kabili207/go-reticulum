package interfaces

import (
	"bytes"
	"testing"
)

func TestHDLCEscapeUnescape_RoundTrip(t *testing.T) {
	t.Parallel()

	cases := [][]byte{
		nil,
		{},
		{0x01, 0x02, 0x03},
		{hdlcFlag},
		{hdlcEsc},
		{0x00, hdlcFlag, 0xFF, hdlcEsc, 0x7E},
	}

	for _, in := range cases {
		enc := hdlcEscape(in)
		out := hdlcUnescape(enc)
		if !bytes.Equal(out, in) {
			t.Fatalf("roundtrip mismatch: in=%x out=%x enc=%x", in, out, enc)
		}
	}
}

