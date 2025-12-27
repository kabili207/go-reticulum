package interfaces

import (
	"errors"
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

type fakeScriptedTarmPort struct {
	mu      sync.Mutex
	reads   []byte
	rdErr   error
	writes  [][]byte
	closed  bool
	closeCh chan struct{}
}

func (f *fakeScriptedTarmPort) Read(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, errors.New("closed")
	}
	if f.rdErr != nil {
		return 0, f.rdErr
	}
	if len(f.reads) == 0 {
		return 0, nil
	}
	p[0] = f.reads[0]
	f.reads = f.reads[1:]
	return 1, nil
}

func (f *fakeScriptedTarmPort) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, errors.New("closed")
	}
	f.writes = append(f.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (f *fakeScriptedTarmPort) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil
	}
	f.closed = true
	if f.closeCh != nil {
		close(f.closeCh)
	}
	return nil
}

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

func TestKISSIntegration_ReadyWithoutPayload_ReleasesQueue(t *testing.T) {
	// Do not run in parallel: overrides global hooks.

	prevOpen := openTarmSerialPort
	prevSleep := kissSleep
	t.Cleanup(func() {
		openTarmSerialPort = prevOpen
		kissSleep = prevSleep
	})
	kissSleep = func(time.Duration) {}

	fp := &fakeScriptedTarmPort{
		reads: []byte{
			kissFEND, kissCmdReady, kissFEND, // READY with no payload
		},
	}
	openTarmSerialPort = func(cfg *serial.Config) (tarmSerialPort, error) {
		return fp, nil
	}

	iface := &Interface{Name: "k", Online: false}
	d := &KISSDriver{
		iface:  iface,
		cfg:    kissConfig{Name: "k", Port: "fake", Speed: 9600, DataBits: 8, HWMTU: 564, ReadTimeout: 10 * time.Millisecond},
		stopCh: make(chan struct{}),
	}

	if err := d.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(d.Close)

	d.sendMu.Lock()
	d.interfaceReady = false
	d.packetQueue = [][]byte{[]byte("hello")}
	d.sendMu.Unlock()

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		fp.mu.Lock()
		nw := len(fp.writes)
		fp.mu.Unlock()
		if nw > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected a write after READY frame")
}

func TestKISSIntegration_PanicOnInterfaceError_ReadError(t *testing.T) {
	// Do not run in parallel: overrides global hooks.

	prevOpen := openTarmSerialPort
	prevSleep := kissSleep
	prevPanic := PanicFunc
	prevPanicOn := PanicOnInterfaceErrorProvider
	t.Cleanup(func() {
		openTarmSerialPort = prevOpen
		kissSleep = prevSleep
		PanicFunc = prevPanic
		PanicOnInterfaceErrorProvider = prevPanicOn
	})
	kissSleep = func(time.Duration) {}

	panicCalled := make(chan struct{}, 1)
	PanicFunc = func() { panicCalled <- struct{}{} }
	PanicOnInterfaceErrorProvider = func() bool { return true }

	fp := &fakeScriptedTarmPort{rdErr: errors.New("boom")}
	openTarmSerialPort = func(cfg *serial.Config) (tarmSerialPort, error) {
		return fp, nil
	}

	iface := &Interface{Name: "k", Online: false}
	d := &KISSDriver{
		iface:  iface,
		cfg:    kissConfig{Name: "k", Port: "fake", Speed: 9600, DataBits: 8, HWMTU: 564, ReadTimeout: 10 * time.Millisecond},
		stopCh: make(chan struct{}),
	}

	if err := d.start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(d.Close)

	select {
	case <-panicCalled:
	case <-time.After(1 * time.Second):
		t.Fatalf("expected PanicFunc to be called on read error")
	}
}
