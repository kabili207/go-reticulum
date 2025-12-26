package rns

import (
	"encoding/hex"
	"sync"
	"testing"
	"time"

	umsgpack "main/rns/vendor"
)

type integrationMessageTest struct {
	ID   string
	Data string
}

func (m *integrationMessageTest) MsgType() uint16 { return 0xABCD }
func (m *integrationMessageTest) Pack() ([]byte, error) {
	return umsgpack.Packb([]any{m.ID, m.Data})
}
func (m *integrationMessageTest) Unpack(raw []byte) error {
	var v []any
	if err := umsgpack.Unpackb(raw, &v); err != nil {
		return err
	}
	if len(v) >= 2 {
		switch vv := v[0].(type) {
		case string:
			m.ID = vv
		case []byte:
			m.ID = string(vv)
		}
		switch vv := v[1].(type) {
		case string:
			m.Data = vv
		case []byte:
			m.Data = string(vv)
		}
	}
	return nil
}

func TestIntegration_ChannelRoundTrip(t *testing.T) {
	requireIntegration(t)
	resetKnownDestinationsForTest()
	withIntegrationTransport(t, func() {
		// Mirrors `tests/link.py::test_10_channel_round_trip`, but uses the in-process transport harness.
		prvHex := "f8953ffaf607627e615603ff1530c82c434cf87c07179dd7689ea776f30b964cfb7ba6164af00c5111a45e69e57d885e1285f8dbfe3a21e95ae17cf676b0f8b7"
		prv, _ := hex.DecodeString(prvHex)
		id, err := IdentityFromBytes(prv)
		if err != nil {
			t.Fatalf("IdentityFromBytes: %v", err)
		}

		const appName = "rns_unit_tests"
		destOut, err := NewDestination(id, DestinationOUT, DestinationSINGLE, appName, "link", "establish")
		if err != nil {
			t.Fatalf("NewDestination(out): %v", err)
		}
		_, err = NewDestination(id, DestinationIN, DestinationSINGLE, appName, "link", "establish")
		if err != nil {
			t.Fatalf("NewDestination(in): %v", err)
		}

		l, err := NewOutgoingLink(destOut, LinkModeDefault, nil, nil)
		if err != nil {
			t.Fatalf("NewOutgoingLink: %v", err)
		}
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if l.Status == LinkActive {
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		if l.Status != LinkActive {
			t.Fatalf("expected link active, got %d", l.Status)
		}
		peer := findPeerLinkTest(l)
		if peer == nil {
			t.Fatalf("expected peer link")
		}

		initCh := l.Channel()
		peerCh := peer.Channel()
		if initCh == nil || peerCh == nil {
			t.Fatalf("channel nil")
		}

		// Register message type and responder-side echo handler.
		if err := peerCh._register_message_type(&integrationMessageTest{}, false); err != nil {
			t.Fatalf("peer register msg type: %v", err)
		}
		if err := initCh._register_message_type(&integrationMessageTest{}, false); err != nil {
			t.Fatalf("init register msg type: %v", err)
		}

		peerCh.AddMessageHandler(func(msg MessageBase) bool {
			mt, ok := msg.(*integrationMessageTest)
			if !ok {
				return false
			}
			resp := &integrationMessageTest{ID: mt.ID, Data: mt.Data + " back"}
			_, _ = peerCh.Send(resp)
			return true
		})

		var (
			mu       sync.Mutex
			received []*integrationMessageTest
		)
		initCh.AddMessageHandler(func(msg MessageBase) bool {
			mt, ok := msg.(*integrationMessageTest)
			if !ok {
				return false
			}
			mu.Lock()
			received = append(received, mt)
			mu.Unlock()
			return true
		})

		testMsg := &integrationMessageTest{ID: "id", Data: "Hello"}
		if _, err := initCh.Send(testMsg); err != nil {
			t.Fatalf("Send: %v", err)
		}

		waitUntil := time.Now().Add(2 * time.Second)
		for time.Now().Before(waitUntil) {
			mu.Lock()
			n := len(received)
			mu.Unlock()
			if n >= 1 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		mu.Lock()
		defer mu.Unlock()
		if len(received) != 1 {
			t.Fatalf("expected 1 message, got %d", len(received))
		}
		rx := received[0]
		if rx.Data != "Hello back" {
			t.Fatalf("unexpected message data: %q", rx.Data)
		}
		if rx.ID != testMsg.ID {
			t.Fatalf("unexpected message id: %q", rx.ID)
		}
		if len(l.Channel().rxRing) != 0 {
			t.Fatalf("expected rx ring empty, got %d", len(l.Channel().rxRing))
		}

		l.Teardown()
	})
}
