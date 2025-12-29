package main

import (
	"os"
	"testing"

	rns "github.com/svanichkin/go-reticulum/rns"
)

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("RNS_INTEGRATION") == "" {
		t.Skip("set RNS_INTEGRATION=1 to run integration tests")
	}
}

func TestIntegration_Broadcast_InboundDeliversToPlainDestination(t *testing.T) {
	requireIntegration(t)

	// Mutates global transport state.
	prevTransportID := rns.TransportIdentity
	t.Cleanup(func() { rns.TransportIdentity = prevTransportID })

	id, err := rns.NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity: %v", err)
	}
	rns.TransportIdentity = id

	dest, err := rns.NewDestination(nil, rns.DestinationIN, rns.DestinationPLAIN, appName, "broadcast", "public_information")
	if err != nil {
		t.Fatalf("NewDestination: %v", err)
	}

	got := make(chan string, 1)
	dest.SetPacketCallback(func(data []byte, _ *rns.Packet) {
		got <- string(data)
	})

	pkt := rns.NewPacket(dest, []byte("hello"))
	if pkt == nil {
		t.Fatalf("NewPacket returned nil")
	}
	if err := pkt.Pack(); err != nil {
		t.Fatalf("Pack: %v", err)
	}

	ifc := &rns.Interface{Name: "if0", IN: true, OUT: true, IngressControl: false}
	rns.Inbound(append([]byte(nil), pkt.Raw...), ifc)

	select {
	case v := <-got:
		if v != "hello" {
			t.Fatalf("expected hello got %q", v)
		}
	default:
		t.Fatalf("expected callback")
	}
}
