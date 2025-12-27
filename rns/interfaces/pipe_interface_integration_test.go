package interfaces

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPipeIntegration_EchoesFramesViaHelperProcess(t *testing.T) {
	// Does not require external binaries; uses the test binary as a helper subprocess.

	for _, a := range os.Args {
		if a == "echo" {
			t.Fatalf("helper process should not run this test")
		}
	}

	oldInbound := InboundHandler
	t.Cleanup(func() { InboundHandler = oldInbound })

	recv := make(chan []byte, 1)
	InboundHandler = func(raw []byte, _ *Interface) {
		cp := append([]byte(nil), raw...)
		select {
		case recv <- cp:
		default:
		}
	}

	// Use an explicit argv string, compatible with our shlexSplit.
	argv0 := shellQuote(os.Args[0])
	argv := argv0 + " -test.run TestPipeHelperProcess -- echo"
	iface, err := NewPipeInterface("p0", map[string]string{
		"command": argv,
	})
	if err != nil {
		t.Fatalf("NewPipeInterface: %v", err)
	}
	t.Cleanup(func() { iface.Detach() })

	// Send payload with bytes that require escaping.
	payload := []byte{0x01, hdlcFlag, 0x02, hdlcEsc, 0x03}
	iface.ProcessOutgoing(payload)

	select {
	case got := <-recv:
		if !bytes.Equal(got, payload) {
			t.Fatalf("unexpected payload: want %x got %x", payload, got)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for inbound payload")
	}
}

func TestPipeHelperProcess(t *testing.T) {
	// Helper process for TestPipeIntegration_*.
	// It reads stdin and writes it to stdout (raw echo) until EOF.

	run := false
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "--" && i+1 < len(os.Args) && os.Args[i+1] == "echo" {
			run = true
			break
		}
	}
	if !run {
		return
	}
	_, _ = io.Copy(os.Stdout, os.Stdin)
	os.Exit(0)
}

func shellQuote(s string) string {
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, " \t\r\n\"") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}
