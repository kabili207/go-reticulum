package interfaces

import (
	"bytes"
	"crypto/sha256"
	"net"
	"testing"
	"time"
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

func TestInterface_ProcessAnnounceRaw_DefaultAnnounceCapProvider(t *testing.T) {
	// Do not run in parallel: overrides global hooks.

	prevCap := DefaultAnnounceCapProvider
	prevHeader := HeaderMinSize
	t.Cleanup(func() {
		DefaultAnnounceCapProvider = prevCap
		HeaderMinSize = prevHeader
	})

	DefaultAnnounceCapProvider = func() float64 { return 1.0 }
	HeaderMinSize = 0

	serverSide, clientSide := net.Pipe()
	defer serverSide.Close()
	defer clientSide.Close()

	iface := &Interface{
		Name:      "local0",
		Type:      "LocalInterface",
		Online:    true,
		Bitrate:   62500,
		AnnounceCap: 0, // should default via provider
	}
	iface.setLocalConn(serverSide)

	raw := []byte("hello-announce")
	iface.ProcessAnnounceRaw(raw, 1)

	_ = clientSide.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 256)
	n, err := clientSide.Read(buf)
	if err != nil || n == 0 {
		t.Fatalf("expected framed bytes written, got n=%d err=%v", n, err)
	}
}

func TestInterface_IncomingAnnounceFrequency_Basic(t *testing.T) {
	t.Parallel()

	iface := &Interface{Name: "if0"}
	iface.ReceivedAnnounce()
	iface.ReceivedAnnounce()

	if v := iface.IncomingAnnounceFrequency(); v <= 0 {
		t.Fatalf("expected >0 frequency, got %f", v)
	}
}

func TestInterface_Hash_MatchesSHA256OfString(t *testing.T) {
	t.Parallel()

	iface := &Interface{Name: "if0"}
	want := sha256.Sum256([]byte("if0"))
	got := iface.Hash()
	if !bytes.Equal(got, want[:]) {
		t.Fatalf("unexpected hash: want %x got %x", want[:], got)
	}
}
