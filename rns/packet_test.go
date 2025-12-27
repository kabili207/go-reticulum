package rns

import (
	"bytes"
	"testing"
)

func TestPacket_PackUnpack_Header1_Announce(t *testing.T) {
	t.Parallel()

	id, err := NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity: %v", err)
	}
	dst, err := NewDestination(id, DestinationIN, DestinationSINGLE, "test", "packet")
	if err != nil {
		t.Fatalf("NewDestination: %v", err)
	}

	p := NewPacket(dst, []byte("announce"), WithPacketType(PacketTypeAnnounce))
	if p == nil {
		t.Fatalf("NewPacket returned nil")
	}
	if err := p.Pack(); err != nil {
		t.Fatalf("Pack: %v", err)
	}

	var q Packet
	q.Raw = p.RawBytes()
	if ok := q.Unpack(); !ok {
		t.Fatalf("Unpack failed")
	}
	if q.HeaderType != HeaderType1 || q.PacketType != PacketTypeAnnounce || q.Context != PacketCtxNone {
		t.Fatalf("unexpected parsed fields: header=%d type=%d ctx=%d", q.HeaderType, q.PacketType, q.Context)
	}
	if !bytes.Equal(q.DestinationHash, p.DestinationHash) {
		t.Fatalf("destination hash mismatch")
	}
	if !bytes.Equal(q.Data, []byte("announce")) {
		t.Fatalf("data mismatch: %q", string(q.Data))
	}
}

func TestPacket_HashIgnoresHops(t *testing.T) {
	t.Parallel()

	dst, err := NewDestination(nil, DestinationOUT, DestinationPLAIN, "test", "packet", "hash")
	if err != nil {
		t.Fatalf("NewDestination: %v", err)
	}

	p1 := NewPacket(dst, []byte("data"))
	if err := p1.Pack(); err != nil {
		t.Fatalf("Pack1: %v", err)
	}
	h1 := p1.GetHash()

	p2 := NewPacket(dst, []byte("data"))
	p2.Hops = 3
	if err := p2.Pack(); err != nil {
		t.Fatalf("Pack2: %v", err)
	}
	h2 := p2.GetHash()

	if !bytes.Equal(h1, h2) {
		t.Fatalf("expected hash to ignore hops")
	}
}

func TestPacket_Header2_TransportID_NotInHash(t *testing.T) {
	t.Parallel()

	dst, err := NewDestination(nil, DestinationOUT, DestinationPLAIN, "test", "packet", "hdr2")
	if err != nil {
		t.Fatalf("NewDestination: %v", err)
	}

	transportA := bytes.Repeat([]byte{0xAA}, ReticulumTruncatedHashLength/8)
	transportB := bytes.Repeat([]byte{0xBB}, ReticulumTruncatedHashLength/8)

	p1 := NewPacket(dst, []byte("a"), WithHeaderType(HeaderType2), WithTransportID(transportA), WithPacketType(PacketTypeAnnounce))
	if err := p1.Pack(); err != nil {
		t.Fatalf("Pack1: %v", err)
	}
	p2 := NewPacket(dst, []byte("a"), WithHeaderType(HeaderType2), WithTransportID(transportB), WithPacketType(PacketTypeAnnounce))
	if err := p2.Pack(); err != nil {
		t.Fatalf("Pack2: %v", err)
	}

	if !bytes.Equal(p1.GetHash(), p2.GetHash()) {
		t.Fatalf("expected header2 hash to ignore transport_id")
	}
}

