package interfaces

import (
	"bytes"
	"io"
	"sync"
	"testing"
	"time"
)

func TestAX25_buildAX25Address_Known(t *testing.T) {
	t.Parallel()

	cfg := ax25kissConfig{
		DstCall: []byte("APZRNS"),
		DstSSID: 0,
		Callsign: []byte("N0CALL"),
		SSID:     0,
	}
	addr := buildAX25Address(cfg)
	if len(addr) != 14 {
		t.Fatalf("unexpected addr length %d", len(addr))
	}
	// Destination callsign (6 chars) left-shifted by 1.
	if addr[0] != byte('A')<<1 || addr[5] != byte('S')<<1 {
		t.Fatalf("unexpected dst callsign encoding: %x", addr[:7])
	}
	// Destination SSID byte includes 0x60 base in this implementation.
	if addr[6] != 0x60|(0<<1) {
		t.Fatalf("unexpected dst ssid byte %02x", addr[6])
	}
	// Source callsign.
	if addr[7] != byte('N')<<1 || addr[12] != byte('L')<<1 {
		t.Fatalf("unexpected src callsign encoding: %x", addr[7:])
	}
	// Source SSID byte includes 0x61 base.
	if addr[13] != 0x61|(0<<1) {
		t.Fatalf("unexpected src ssid byte %02x", addr[13])
	}
}

func TestAX25_buildFrame_EscapesKISSBytes(t *testing.T) {
	t.Parallel()

	cfg := ax25kissConfig{
		HWMTU: 564,
	}
	d := &AX25KISSDriver{
		cfg:        cfg,
		addrHeader: buildAX25Address(ax25kissConfig{DstCall: []byte("APZRNS"), DstSSID: 0, Callsign: []byte("N0CALL"), SSID: 0}),
	}

	// Payload includes bytes that must be escaped inside KISS framing.
	payload := []byte{0x01, kissFEND, 0x02, kissFESC, 0x03}
	frame := d.buildFrame(payload)
	if len(frame) < 4 {
		t.Fatalf("unexpected frame length %d", len(frame))
	}
	if frame[0] != kissFEND || frame[1] != kissCmdData || frame[len(frame)-1] != kissFEND {
		t.Fatalf("unexpected frame envelope: %x", frame[:min(16, len(frame))])
	}

	inner := frame[2 : len(frame)-1]
	// Ensure no raw FEND exists inside payload (it should be escaped).
	if bytes.Contains(inner, []byte{kissFEND}) {
		t.Fatalf("raw FEND present in escaped inner frame: %x", inner)
	}
}

func TestAX25_ProcessOutgoing_QueuesWhenNotReady(t *testing.T) {
	t.Parallel()

	iface := &Interface{Online: true}
	d := &AX25KISSDriver{
		iface:          iface,
		cfg:            ax25kissConfig{HWMTU: 4},
		interfaceReady: false,
	}
	d.ProcessOutgoing([]byte{1, 2, 3, 4, 5})
	if len(d.packetQueue) != 1 {
		t.Fatalf("expected queued packet, got %d", len(d.packetQueue))
	}
	if got := d.packetQueue[0]; len(got) != 4 {
		t.Fatalf("expected truncation to HWMTU, got %d bytes", len(got))
	}
}

type fakeByteSerial struct {
	mu     sync.Mutex
	reads  []byte
	ridx   int
	writes [][]byte
	closed bool
}

func (f *fakeByteSerial) Read(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, io.EOF
	}
	if f.ridx >= len(f.reads) {
		f.closed = true
		return 0, io.EOF
	}
	p[0] = f.reads[f.ridx]
	f.ridx++
	return 1, nil
}

func (f *fakeByteSerial) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writes = append(f.writes, append([]byte(nil), p...))
	return len(p), nil
}

func (f *fakeByteSerial) Close() error {
	f.mu.Lock()
	f.closed = true
	f.mu.Unlock()
	return nil
}

func TestAX25_FlowControl_READYFlushesQueue(t *testing.T) {
	// Do not run in parallel: modifies global sleep hook.

	prevSleep := ax25Sleep
	t.Cleanup(func() { ax25Sleep = prevSleep })
	ax25Sleep = func(time.Duration) {}

	iface := &Interface{Online: true}
	fake := &fakeByteSerial{
		// Emit a READY frame: FEND, cmd=0x0F, value=0x01, then EOF.
		reads: []byte{kissFEND, kissCmdReady, 0x01},
	}

	d := &AX25KISSDriver{
		iface:          iface,
		cfg:            ax25kissConfig{HWMTU: 64, FlowControl: true, FlowTimeout: 50 * time.Millisecond, FrameTimeout: 10 * time.Millisecond},
		serial:         fake,
		interfaceReady: true,
		stopCh:         make(chan struct{}),
		addrHeader:     buildAX25Address(ax25kissConfig{DstCall: []byte("APZRNS"), DstSSID: 0, Callsign: []byte("N0CALL"), SSID: 0}),
	}

	// Send two packets; second should queue until READY is read.
	d.ProcessOutgoing([]byte("one"))
	d.ProcessOutgoing([]byte("two"))
	if len(d.packetQueue) != 1 {
		t.Fatalf("expected 1 queued packet, got %d", len(d.packetQueue))
	}

	go d.readLoop()

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		fake.mu.Lock()
		nWrites := len(fake.writes)
		fake.mu.Unlock()
		// One outgoing frame for "one", one for "two", plus configure commands are not sent here.
		if nWrites >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	close(d.stopCh)

	fake.mu.Lock()
	nWrites := len(fake.writes)
	fake.mu.Unlock()
	if nWrites < 2 {
		t.Fatalf("expected >=2 frame writes, got %d", nWrites)
	}
	if len(d.packetQueue) != 0 {
		t.Fatalf("expected queue drained, got %d", len(d.packetQueue))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
