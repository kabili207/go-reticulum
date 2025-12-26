//go:build windows

package ble

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"tinygo.org/x/bluetooth"
)

type transport struct {
	targetName string
	targetAddr string

	mu     sync.Mutex
	device bluetooth.Device
	rxChr  bluetooth.DeviceCharacteristic
	txChr  bluetooth.DeviceCharacteristic

	connected     bool
	disappeared   bool
	reconnectCh   chan struct{}
	reconnectOnce sync.Once
	txOnce        sync.Once

	rxMu  sync.Mutex
	rxBuf bytes.Buffer
	rxCh  chan struct{}

	txMu    sync.Mutex
	txQueue []byte
	txCond  *sync.Cond

	stopCh chan struct{}
	wg     sync.WaitGroup
}

const (
	scanTimeout    = 2 * time.Second
	connectTimeout = 5 * time.Second
)

var (
	nusServiceUUID = bluetooth.ServiceUUIDNordicUART
	nusRXUUID      = bluetooth.CharacteristicUUIDUARTRX
	nusTXUUID      = bluetooth.CharacteristicUUIDUARTTX
)

var (
	winOnce   sync.Once
	winErr    error
	winAdapter = bluetooth.DefaultAdapter
)

func ensureDevice() error {
	winOnce.Do(func() { winErr = winAdapter.Enable() })
	return winErr
}

func New(target string) (Transport, error) {
	if err := ensureDevice(); err != nil {
		return nil, err
	}
	name, addr := ParseTarget(target)
	t := &transport{
		targetName:  name,
		targetAddr:  addr,
		stopCh:      make(chan struct{}),
		rxCh:        make(chan struct{}, 1),
		reconnectCh: make(chan struct{}, 1),
	}
	t.txCond = sync.NewCond(&t.txMu)
	return t, nil
}

func (b *transport) String() string {
	if b.targetAddr != "" {
		return "ble://" + b.targetAddr
	}
	if b.targetName != "" {
		return "ble://" + b.targetName
	}
	return "ble://"
}

func (b *transport) IsOpen() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.connected
}

func (b *transport) DeviceDisappeared() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.disappeared
}

func (b *transport) Open(ctx context.Context) error {
	if err := ensureDevice(); err != nil {
		return err
	}
	b.mu.Lock()
	if b.connected {
		b.mu.Unlock()
		return nil
	}
	b.mu.Unlock()

	b.reconnectOnce.Do(func() {
		b.wg.Add(1)
		go func() { defer b.wg.Done(); b.reconnectLoop() }()
	})
	if err := b.connectOnce(ctx); err != nil {
		return err
	}
	b.txOnce.Do(func() {
		b.wg.Add(1)
		go func() { defer b.wg.Done(); b.txLoop() }()
	})
	return nil
}

func (b *transport) connectOnce(ctx context.Context) error {
	type found struct{ addr bluetooth.Address }
	foundCh := make(chan found, 1)

	go func() {
		_ = winAdapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			if !result.AdvertisementPayload.HasServiceUUID(nusServiceUUID) {
				return
			}
			localName := result.LocalName()

			if b.targetName == "" && b.targetAddr == "" {
				if !strings.HasPrefix(localName, "RNode ") {
					return
				}
			} else {
				if b.targetAddr != "" && !strings.EqualFold(result.Address.String(), b.targetAddr) {
					return
				}
				if b.targetName != "" && localName != b.targetName {
					return
				}
			}
			select { case foundCh <- found{addr: result.Address}: default: }
			_ = adapter.StopScan()
		})
	}()

	var f found
	select {
	case <-ctx.Done():
		_ = winAdapter.StopScan()
		return ctx.Err()
	case <-time.After(scanTimeout):
		_ = winAdapter.StopScan()
		return errors.New("ble scan timeout")
	case f = <-foundCh:
	}

	_ = connectTimeout
	device, err := winAdapter.Connect(f.addr, bluetooth.ConnectionParams{})
	if err != nil {
		return err
	}

	services, err := device.DiscoverServices([]bluetooth.UUID{nusServiceUUID})
	if err != nil || len(services) == 0 {
		_ = device.Disconnect()
		if err == nil {
			err = errors.New("nus service not found")
		}
		return err
	}
	svc := services[0]
	chars, err := svc.DiscoverCharacteristics([]bluetooth.UUID{nusRXUUID, nusTXUUID})
	if err != nil || len(chars) < 2 {
		_ = device.Disconnect()
		if err == nil {
			err = errors.New("nus characteristics not found")
		}
		return err
	}
	var rx, tx bluetooth.DeviceCharacteristic
	for _, c := range chars {
		switch {
		case c.UUID() == nusRXUUID:
			rx = c
		case c.UUID() == nusTXUUID:
			tx = c
		}
	}
	if err := tx.EnableNotifications(func(value []byte) {
		b.rxMu.Lock()
		b.rxBuf.Write(value)
		b.rxMu.Unlock()
		select { case b.rxCh <- struct{}{}: default: }
	}); err != nil {
		_ = device.Disconnect()
		return err
	}

	b.mu.Lock()
	b.device = device
	b.rxChr = rx
	b.txChr = tx
	b.connected = true
	b.disappeared = false
	b.mu.Unlock()

	select { case b.reconnectCh <- struct{}{}: default: }
	return nil
}

func (b *transport) reconnectLoop() {
	for {
		select {
		case <-b.stopCh:
			return
		case <-b.reconnectCh:
		default:
			time.Sleep(250 * time.Millisecond)
		}
		for {
			select {
			case <-b.stopCh:
				return
			default:
			}
			if b.IsOpen() {
				break
			}
			_ = b.connectOnce(context.Background())
			if b.IsOpen() {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (b *transport) Close() error {
	b.mu.Lock()
	device := b.device
	b.connected = false
	b.mu.Unlock()

	select { case <-b.stopCh: default: close(b.stopCh) }
	if b.txCond != nil {
		b.txCond.Broadcast()
	}
	_ = device.Disconnect()
	b.wg.Wait()
	return nil
}

func (b *transport) Read(p []byte) (int, error) {
	for {
		b.rxMu.Lock()
		if b.rxBuf.Len() > 0 {
			n, _ := b.rxBuf.Read(p)
			b.rxMu.Unlock()
			return n, nil
		}
		b.rxMu.Unlock()
		if !b.IsOpen() {
			return 0, io.EOF
		}
		select {
		case <-b.rxCh:
			continue
		case <-b.stopCh:
			return 0, io.EOF
		case <-time.After(250 * time.Millisecond):
			return 0, nil
		}
	}
}

func (b *transport) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	b.mu.Lock()
	rx := b.rxChr
	connected := b.connected
	b.mu.Unlock()
	if !connected {
		b.txMu.Lock()
		b.txQueue = append(b.txQueue, p...)
		if b.txCond != nil {
			b.txCond.Signal()
		}
		b.txMu.Unlock()
		return len(p), nil
	}
	for _, chunk := range ChunkByMaxWriteNoRsp(p, 20) {
		_, _ = rx.WriteWithoutResponse(chunk)
	}
	return len(p), nil
}

func (b *transport) txLoop() {
	for {
		select {
		case <-b.stopCh:
			return
		default:
		}

		b.mu.Lock()
		rx := b.rxChr
		connected := b.connected
		b.mu.Unlock()
		if !connected {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		b.txMu.Lock()
		for len(b.txQueue) == 0 {
			if b.txCond != nil {
				b.txCond.Wait()
			} else {
				b.txMu.Unlock()
				time.Sleep(100 * time.Millisecond)
				b.txMu.Lock()
			}
			select {
			case <-b.stopCh:
				b.txMu.Unlock()
				return
			default:
			}
		}
		buf := b.txQueue
		b.txQueue = nil
		b.txMu.Unlock()

		for _, chunk := range ChunkByMaxWriteNoRsp(buf, 20) {
			_, _ = rx.WriteWithoutResponse(chunk)
		}
	}
}
