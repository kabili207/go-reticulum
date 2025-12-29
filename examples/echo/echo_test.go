package main

import (
	"bytes"
	"testing"
	"time"

	rns "github.com/svanichkin/go-reticulum/rns"
)

type miniTransport struct{}

func (miniTransport) GetFirstHopTimeout([]byte) time.Duration { return 10 * time.Millisecond }
func (miniTransport) HopsTo([]byte) int                       { return 1 }
func (miniTransport) GetPacketRSSI([]byte) *float64           { return nil }
func (miniTransport) GetPacketSNR([]byte) *float64            { return nil }
func (miniTransport) GetPacketQ([]byte) *float64              { return nil }

func findDestinationByHash(hash []byte) *rns.Destination {
	var fallback *rns.Destination
	for _, d := range rns.Destinations {
		if d == nil || len(d.Hash()) == 0 {
			continue
		}
		if bytes.Equal(d.Hash(), hash) {
			if d.Direction == rns.DestinationIN {
				return d
			}
			if fallback == nil {
				fallback = d
			}
		}
	}
	return fallback
}

func (miniTransport) Outbound(p *rns.Packet) bool {
	if p == nil {
		return false
	}
	if !p.Packed {
		_ = p.Pack()
	}
	p.Sent = true
	p.SentAt = time.Now()
	if p.CreateReceipt {
		rc := rns.NewPacketReceipt(p)
		p.Receipt = rc
		rns.Receipts = append(rns.Receipts, rc)
	}

	in := rns.NewPacket(nil, p.Raw)
	if in == nil || !in.Unpack() {
		return false
	}

	// Proofs should be validated against matching receipts.
	if in.PacketType == rns.PacketTypeProof {
		for _, rc := range rns.Receipts {
			if rc == nil || rc.Status != rns.ReceiptSent {
				continue
			}
			if rc.ValidateProofPacket(in) {
				return true
			}
		}
	}

	if d := findDestinationByHash(in.DestinationHash); d != nil {
		in.Destination = d
		_ = d.Receive(in)
		return true
	}
	return true
}

func TestEchoExample_RequestGetsProof(t *testing.T) {
	prevTransport := rns.Transport
	prevDestinations := rns.Destinations
	prevReceipts := rns.Receipts
	t.Cleanup(func() {
		rns.Transport = prevTransport
		rns.Destinations = prevDestinations
		rns.Receipts = prevReceipts
	})

	rns.Transport = miniTransport{}
	rns.Destinations = nil
	rns.Receipts = nil

	serverID, err := rns.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity(server): %v", err)
	}
	serverDest, err := rns.NewDestination(serverID, rns.DestinationIN, rns.DestinationSINGLE, appName, "echo", "request")
	if err != nil {
		t.Fatalf("NewDestination(server): %v", err)
	}
	_ = serverDest.SetProofStrategy(rns.DestinationPROVE_ALL)

	clientDest, err := rns.NewDestination(serverID, rns.DestinationOUT, rns.DestinationSINGLE, appName, "echo", "request")
	if err != nil {
		t.Fatalf("NewDestination(client): %v", err)
	}

	req := rns.NewPacket(clientDest, rns.IdentityGetRandomHash())
	if req == nil {
		t.Fatalf("NewPacket returned nil")
	}
	rc := req.Send()
	if rc == nil {
		t.Fatalf("Send returned nil receipt")
	}

	// Wait for proof delivery.
	waitUntil := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(waitUntil) {
		if rc.Status == rns.ReceiptDelivered {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if rc.Status != rns.ReceiptDelivered {
		t.Fatalf("expected receipt delivered, got %d", rc.Status)
	}
	if rc.GetRTT() <= 0 {
		t.Fatalf("expected non-zero RTT")
	}
}
