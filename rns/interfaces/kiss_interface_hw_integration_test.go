package interfaces

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tarm/serial"
)

func TestKISSHW_FlowControlReady_ReleasesQueue(t *testing.T) {
	// Requires a pair of connected serial ports (null-modem / virtual pair).
	// Example (Linux): socat -d -d pty,raw,echo=0 pty,raw,echo=0
	// Then run:
	// RUN_KISS_HW_INTEGRATION=1 KISS_PORT_A=/dev/pts/X KISS_PORT_B=/dev/pts/Y go test ./rns/interfaces -run TestKISSHW_ -count=1

	if os.Getenv("RUN_KISS_HW_INTEGRATION") != "1" {
		t.Skip("set RUN_KISS_HW_INTEGRATION=1 to run")
	}
	portA := strings.TrimSpace(os.Getenv("KISS_PORT_A"))
	portB := strings.TrimSpace(os.Getenv("KISS_PORT_B"))
	if portA == "" || portB == "" {
		t.Skip("set KISS_PORT_A and KISS_PORT_B to run")
	}

	speed := 9600
	if v := strings.TrimSpace(os.Getenv("KISS_SPEED")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			speed = n
		}
	}

	// Start the KISS interface on port A.
	kv := map[string]string{
		"port":                 portA,
		"speed":                fmt.Sprintf("%d", speed),
		"flow_control":         "true",
		"flow_control_timeout": "2",
		"read_timeout":         "50",
	}

	ifc, err := NewKISSInterface("kiss-hw", kv)
	if err != nil {
		t.Fatalf("NewKISSInterface: %v", err)
	}
	t.Cleanup(func() { ifc.Detach() })

	// Open the peer side on port B.
	peer, err := serial.OpenPort(&serial.Config{
		Name:        portB,
		Baud:        speed,
		Size:        8,
		StopBits:    serial.Stop1,
		Parity:      serial.ParityNone,
		ReadTimeout: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("OpenPort peer: %v", err)
	}
	t.Cleanup(func() { _ = peer.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Send two frames; second should queue until READY is received.
	payload1 := []byte("hello-1")
	payload2 := []byte("hello-2")
	ifc.ProcessOutgoing(payload1)
	ifc.ProcessOutgoing(payload2)

	// From peer, read frames and, after first data frame, send READY without payload.
	var got [][]byte
	readySent := false
	for len(got) < 2 {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for 2 frames, got %d", len(got))
		default:
		}

		frame, err := readKISSFrame(ctx, peer)
		if err != nil {
			continue
		}
		if frame.cmd != kissCmdData {
			continue
		}
		got = append(got, frame.payload)
		if len(got) == 1 && !readySent {
			// READY frame without payload: FEND CMD_READY FEND (our parser should tolerate this).
			if _, err := peer.Write([]byte{kissFEND, kissCmdReady, kissFEND}); err != nil {
				t.Fatalf("peer write READY: %v", err)
			}
			readySent = true
		}
	}

	if !bytes.Equal(got[0], payload1) || !bytes.Equal(got[1], payload2) {
		t.Fatalf("unexpected payloads: got0=%q got1=%q", string(got[0]), string(got[1]))
	}
}

func TestKISSHW_Beacon_PaddedTo15(t *testing.T) {
	// Requires a pair of connected serial ports (null-modem / virtual pair).
	if os.Getenv("RUN_KISS_HW_INTEGRATION") != "1" {
		t.Skip("set RUN_KISS_HW_INTEGRATION=1 to run")
	}
	portA := strings.TrimSpace(os.Getenv("KISS_PORT_A"))
	portB := strings.TrimSpace(os.Getenv("KISS_PORT_B"))
	if portA == "" || portB == "" {
		t.Skip("set KISS_PORT_A and KISS_PORT_B to run")
	}

	speed := 9600
	if v := strings.TrimSpace(os.Getenv("KISS_SPEED")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			speed = n
		}
	}

	callsign := "TESTBEACON"
	kv := map[string]string{
		"port":         portA,
		"speed":        fmt.Sprintf("%d", speed),
		"id_interval":  "1",
		"id_callsign":  callsign,
		"read_timeout": "50",
	}

	ifc, err := NewKISSInterface("kiss-hw", kv)
	if err != nil {
		t.Fatalf("NewKISSInterface: %v", err)
	}
	t.Cleanup(func() { ifc.Detach() })

	peer, err := serial.OpenPort(&serial.Config{
		Name:        portB,
		Baud:        speed,
		Size:        8,
		StopBits:    serial.Stop1,
		Parity:      serial.ParityNone,
		ReadTimeout: 50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("OpenPort peer: %v", err)
	}
	t.Cleanup(func() { _ = peer.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Start beacon timer by sending a non-beacon frame.
	ifc.ProcessOutgoing([]byte("start"))

	// Expect beacon frame within ~1 interval after first tx.
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("timeout waiting for beacon frame")
		default:
		}
		frame, err := readKISSFrame(ctx, peer)
		if err != nil {
			continue
		}
		if frame.cmd != kissCmdData {
			continue
		}
		if !bytes.HasPrefix(frame.payload, []byte(callsign)) {
			continue
		}
		if len(frame.payload) < 15 {
			t.Fatalf("expected beacon payload >=15 bytes, got %d", len(frame.payload))
		}
		return
	}
}

type kissFrame struct {
	cmd     byte
	payload []byte
}

func readKISSFrame(ctx context.Context, port *serial.Port) (kissFrame, error) {
	// Minimal KISS reader for integration tests.
	var (
		inFrame bool
		escape  bool
		cmd     byte = kissCmdUnknown
		buf          = make([]byte, 0, 2048)
	)

	tmp := []byte{0}
	for {
		select {
		case <-ctx.Done():
			return kissFrame{}, ctx.Err()
		default:
		}
		n, err := port.Read(tmp)
		if err != nil {
			return kissFrame{}, err
		}
		if n == 0 {
			continue
		}
		b := tmp[0]

		if inFrame && b == kissFEND && cmd != kissCmdUnknown {
			return kissFrame{cmd: cmd, payload: append([]byte(nil), buf...)}, nil
		}
		if b == kissFEND {
			inFrame = true
			escape = false
			cmd = kissCmdUnknown
			buf = buf[:0]
			continue
		}
		if !inFrame {
			continue
		}
		if len(buf) == 0 && cmd == kissCmdUnknown {
			cmd = b & 0x0F
			continue
		}
		if cmd != kissCmdData {
			// ignore non-data payload bytes
			continue
		}
		if b == kissFESC {
			escape = true
			continue
		}
		if escape {
			if b == kissTFEND {
				b = kissFEND
			} else if b == kissTFESC {
				b = kissFESC
			}
			escape = false
		}
		buf = append(buf, b)
	}
}

