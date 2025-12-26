package interfaces

import (
	"io"
	"sync"
	"testing"

	"github.com/tarm/serial"
)

type fakeTarmPort struct {
	mu     sync.Mutex
	writes [][]byte
}

func (f *fakeTarmPort) Read(p []byte) (int, error) { return 0, io.EOF }
func (f *fakeTarmPort) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writes = append(f.writes, append([]byte(nil), p...))
	return len(p), nil
}
func (f *fakeTarmPort) Close() error { return nil }

func TestTarmSerialSeam_OpenHookIsUsed(t *testing.T) {
	// Do not run in parallel: modifies global hook.

	prev := openTarmSerialPort
	t.Cleanup(func() { openTarmSerialPort = prev })

	fp := &fakeTarmPort{}
	openTarmSerialPort = func(cfg *serial.Config) (tarmSerialPort, error) {
		if cfg == nil || cfg.Name == "" {
			t.Fatalf("unexpected cfg: %#v", cfg)
		}
		return fp, nil
	}

	d := &AX25KISSDriver{cfg: ax25kissConfig{Port: "fake", Speed: 9600, DataBits: 8}}
	if err := d.openPort(); err != nil {
		t.Fatalf("openPort: %v", err)
	}
	if d.getSerial() == nil {
		t.Fatalf("expected serial to be set")
	}
}

