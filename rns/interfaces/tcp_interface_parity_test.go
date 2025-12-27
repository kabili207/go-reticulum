package interfaces

import (
	"net"
	"testing"
)

type testSigner struct{}

func (testSigner) Sign([]byte) ([]byte, error) { return []byte("sig"), nil }

func TestTCPClientInterface_OptimiseMTU(t *testing.T) {
	ci := &TCPClientInterface{AutoconfigureMTU: true}
	ci.Bitrate = 10_000_001
	ci.OptimiseMTU()
	if ci.HWMTU != 16384 {
		t.Fatalf("unexpected mtu: %d", ci.HWMTU)
	}
}

func TestTCPServerInterface_SpawnClient_InheritsIFAC(t *testing.T) {
	// Do not run in parallel: overrides global hook.
	old := TCPIFACDeriver
	t.Cleanup(func() { TCPIFACDeriver = old })

	TCPIFACDeriver = func(_, _ string) ([]byte, interface{ Sign([]byte) ([]byte, error) }, []byte, error) {
		return []byte("k"), testSigner{}, []byte("s"), nil
	}

	s, err := NewTCPServerFromConfig(nil, nil, TCPServerConfig{
		Name:        "s0",
		ListenIP:    "127.0.0.1",
		ListenPort:  4242,
		IFACNetname: "n",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	serverSide, clientSide := net.Pipe()
	t.Cleanup(func() { _ = serverSide.Close() })
	t.Cleanup(func() { _ = clientSide.Close() })

	ci := s.spawnClient(serverSide, &net.TCPAddr{IP: net.IPv4(1, 2, 3, 4), Port: 1234})
	if string(ci.IFACKey) != "k" || string(ci.IFACSignature) != "s" || ci.IFACIdentity == nil {
		t.Fatalf("client did not inherit IFAC")
	}
}

