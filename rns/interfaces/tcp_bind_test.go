package interfaces

import (
	"strings"
	"testing"
)

func TestTCPBind_AddressForHost_LiteralIPv4(t *testing.T) {
	addr, err := tcpAddressForHost("127.0.0.1", 1234, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if addr != "127.0.0.1:1234" {
		t.Fatalf("unexpected addr: %q", addr)
	}
}

func TestTCPBind_AddressForHost_LiteralIPv6(t *testing.T) {
	addr, err := tcpAddressForHost("2001:db8::1", 1234, true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if addr != "[2001:db8::1]:1234" {
		t.Fatalf("unexpected addr: %q", addr)
	}
}

func TestTCPBind_AddressForInterface_Validates(t *testing.T) {
	_, err := tcpAddressForInterface("", 1234, false)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}
