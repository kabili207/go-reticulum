package rns

import "testing"

func TestTransportConstants_PythonParity(t *testing.T) {
	maybeParallel(t)

	if ReachabilityUnreachable != 0x00 || ReachabilityDirect != 0x01 || ReachabilityTransport != 0x02 {
		t.Fatalf("unexpected reachability constants")
	}
	if TransportStateUnknown != 0x00 || TransportStateUnresponsive != 0x01 || TransportStateResponsive != 0x02 {
		t.Fatalf("unexpected transport state constants")
	}
	if TransportAppName != "rnstransport" {
		t.Fatalf("unexpected TransportAppName %q", TransportAppName)
	}
}
