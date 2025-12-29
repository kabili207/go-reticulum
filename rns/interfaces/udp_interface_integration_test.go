package interfaces

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time"
)

func freeUDPPort(t *testing.T) int {
	t.Helper()
	c, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("UDP bind not permitted in sandbox: %v", err)
		}
		t.Fatalf("ListenUDP: %v", err)
	}
	port := c.LocalAddr().(*net.UDPAddr).Port
	_ = c.Close()
	return port
}

func TestUDPIntegration_StartUDP_ReceivesInbound(t *testing.T) {
	// Do not run in parallel: overrides global hook.
	old := InboundHandler
	t.Cleanup(func() { InboundHandler = old })

	got := make(chan []byte, 1)
	InboundHandler = func(raw []byte, _ *Interface) { got <- append([]byte(nil), raw...) }

	var ifc Interface
	ifc.HWMTU = 1064
	listenPort := freeUDPPort(t)
	if err := ifc.ConfigureUDP("127.0.0.1", listenPort, "127.0.0.1", listenPort); err != nil {
		t.Fatalf("ConfigureUDP: %v", err)
	}
	if err := ifc.StartUDP(); err != nil {
		t.Fatalf("StartUDP: %v", err)
	}
	t.Cleanup(func() { ifc.StopUDP() })

	c, err := net.DialUDP("udp", nil, ifc.udpBindAddr)
	if err != nil {
		t.Fatalf("DialUDP: %v", err)
	}
	defer c.Close()

	want := []byte("hello")
	if _, err := c.Write(want); err != nil {
		t.Fatalf("write: %v", err)
	}

	select {
	case b := <-got:
		if !bytes.Equal(b, want) {
			t.Fatalf("unexpected inbound: %q", string(b))
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting inbound")
	}
}

func TestUDPIntegration_ProcessOutgoing_SendsToForward(t *testing.T) {
	recv, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("UDP bind not permitted in sandbox: %v", err)
		}
		t.Fatalf("ListenUDP: %v", err)
	}
	defer recv.Close()

	var ifc Interface
	listenPort := freeUDPPort(t)
	if err := ifc.ConfigureUDP("127.0.0.1", listenPort, "127.0.0.1", recv.LocalAddr().(*net.UDPAddr).Port); err != nil {
		t.Fatalf("ConfigureUDP: %v", err)
	}
	if err := ifc.StartUDP(); err != nil {
		t.Fatalf("StartUDP: %v", err)
	}
	t.Cleanup(func() { ifc.StopUDP() })

	want := []byte("payload")
	ifc.udpProcessOutgoing(want)

	_ = recv.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 2048)
	n, _, err := recv.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("ReadFromUDP: %v", err)
	}
	if !bytes.Equal(buf[:n], want) {
		t.Fatalf("unexpected forwarded: %q", string(buf[:n]))
	}
}

func TestUDPIntegration_ProcessOutgoing_WorksWithoutListener(t *testing.T) {
	recv, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("UDP bind not permitted in sandbox: %v", err)
		}
		t.Fatalf("ListenUDP: %v", err)
	}
	defer recv.Close()

	var ifc Interface
	if err := ifc.ConfigureUDP("", 0, "127.0.0.1", recv.LocalAddr().(*net.UDPAddr).Port); err != nil {
		t.Fatalf("ConfigureUDP: %v", err)
	}
	// No StartUDP() here on purpose.

	want := []byte("payload")
	ifc.udpProcessOutgoing(want)

	_ = recv.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 2048)
	n, _, err := recv.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("ReadFromUDP: %v", err)
	}
	if !bytes.Equal(buf[:n], want) {
		t.Fatalf("unexpected forwarded: %q", string(buf[:n]))
	}
}
