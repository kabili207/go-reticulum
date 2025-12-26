package interfaces

import (
	"sync"
	"testing"
	"time"

	"github.com/tarm/serial"
)

type fakeIdleTarmPort struct {
	mu     sync.Mutex
	writes [][]byte
}

func (f *fakeIdleTarmPort) Read(p []byte) (int, error) { return 0, nil }
func (f *fakeIdleTarmPort) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writes = append(f.writes, append([]byte(nil), p...))
	return len(p), nil
}
func (f *fakeIdleTarmPort) Close() error { return nil }

func TestKISS_Start_ConfiguresDeviceWithoutRealSleep(t *testing.T) {
	// Do not run in parallel: overrides global hooks.

	prevOpen := openTarmSerialPort
	prevSleep := kissSleep
	t.Cleanup(func() {
		openTarmSerialPort = prevOpen
		kissSleep = prevSleep
	})

	fp := &fakeIdleTarmPort{}
	openTarmSerialPort = func(cfg *serial.Config) (tarmSerialPort, error) {
		return fp, nil
	}
	sleepCalls := 0
	kissSleep = func(time.Duration) { sleepCalls++ }

	iface := &Interface{Name: "k", Online: false}
	d := &KISSDriver{
		iface:  iface,
		cfg:    kissConfig{Name: "k", Port: "fake", Speed: 9600, DataBits: 8, HWMTU: 564, ReadTimeout: 100 * time.Millisecond, Preamble: 350, TxTail: 20, Persistence: 64, SlotTime: 20},
		stopCh: make(chan struct{}),
	}

	if err := d.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	// Ensure we used the hook instead of sleeping.
	if sleepCalls < 1 {
		t.Fatalf("expected kissSleep to be called")
	}

	// Expect 5 config commands: TXDELAY, TXTAIL, P, SLOTTIME, READY.
	fp.mu.Lock()
	writes := append([][]byte(nil), fp.writes...)
	fp.mu.Unlock()
	if len(writes) < 5 {
		t.Fatalf("expected >=5 config writes, got %d", len(writes))
	}
	wantCmds := []byte{kissCmdTxDelay, kissCmdTxTail, kissCmdP, kissCmdSlot, kissCmdReady}
	for i, cmd := range wantCmds {
		frame := writes[i]
		if len(frame) != 4 || frame[0] != kissFEND || frame[1] != cmd || frame[3] != kissFEND {
			t.Fatalf("unexpected cmd frame %d: %x", i, frame)
		}
	}

	d.Close()
}
