package rns

import (
	"bytes"
	"sync/atomic"
	"testing"
	"time"
)

type stuckOutlet struct {
	name string
	mdu  int
	rtt  float64
	id   atomic.Uint64
}

func (o *stuckOutlet) Send(raw []byte) any {
	_ = raw
	return o.id.Add(1)
}
func (o *stuckOutlet) Resend(packet any) any { return packet }
func (o *stuckOutlet) Mdu() int              { return o.mdu }
func (o *stuckOutlet) Rtt() float64          { return o.rtt }
func (o *stuckOutlet) IsUsable() bool        { return true }
func (o *stuckOutlet) TimedOut()             {}
func (o *stuckOutlet) String() string        { return o.name }
func (o *stuckOutlet) GetPacketID(packet any) any {
	return packet
}
func (o *stuckOutlet) GetPacketState(packet any) MessageState {
	_ = packet
	// Never delivered, so channel window will fill.
	return MSGSTATE_SENT
}
func (o *stuckOutlet) SetPacketTimeoutCallback(packet any, cb func(any), timeout *float64) {
	_ = packet
	_ = cb
	_ = timeout
}
func (o *stuckOutlet) SetPacketDeliveredCallback(packet any, cb func(any)) {
	_ = packet
	_ = cb
}

func TestBufferIntegration_RawChannelWriter_NonBlockingOnLinkNotReady(t *testing.T) {
	o := &stuckOutlet{name: "stuck", mdu: 65535, rtt: 0.05}
	ch := NewChannel(o)

	writer := NewRawChannelWriter(1, ch)

	// First writes should succeed until the channel window is full.
	n1, err := writer.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("write1 err: %v", err)
	}
	if n1 == 0 {
		t.Fatalf("write1 expected progress")
	}

	n2, err := writer.Write([]byte("world"))
	if err != nil {
		t.Fatalf("write2 err: %v", err)
	}
	if n2 == 0 {
		t.Fatalf("write2 expected progress")
	}

	// Third write should hit ME_LINK_NOT_READY and return 0 (non-blocking).
	start := time.Now()
	n3, err := writer.Write([]byte("!"))
	if err != nil {
		t.Fatalf("write3 err: %v", err)
	}
	if n3 != 0 {
		t.Fatalf("write3 expected 0, got %d", n3)
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Fatalf("write3 unexpectedly blocked")
	}
}

func TestBufferIntegration_ChannelBufferedWriter_FlushBlocksUntilProgress(t *testing.T) {
	// This is a smoke test that Flush uses the blocking write path and that the
	// receiving side can unpack StreamDataMessage.
	oa := newLoopOutlet("a", 65535)
	ob := newLoopOutlet("b", 65535)
	ca := NewChannel(oa)
	cb := NewChannel(ob)
	oa.peer.Store(cb)
	ob.peer.Store(ca)

	reader := NewRawChannelReader(1, cb)
	defer reader.Close()
	writer := CreateWriter(1, ca)
	defer writer.Close()

	payload := bytes.Repeat([]byte("A"), 10000)
	if _, err := writer.Write(payload); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := writer.Flush(); err != nil {
		t.Fatalf("flush: %v", err)
	}
}
