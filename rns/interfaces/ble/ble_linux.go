//go:build linux

package ble

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/examples/lib/dev"
)

type transport struct {
	targetName string
	targetAddr string

	mu    sync.Mutex
	cln   ble.Client
	rxChr *ble.Characteristic
	txChr *ble.Characteristic

	connected     bool
	disappeared   bool
	reconnectCh   chan struct{}
	reconnectOnce sync.Once
	txOnce        sync.Once

	rxMu    sync.Mutex
	rxBuf   bytes.Buffer
	rxCh    chan struct{}
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
	uartServiceUUID = ble.MustParse("6E400001-B5A3-F393-E0A9-E50E24DCCA9E")
	uartRXCharUUID  = ble.MustParse("6E400002-B5A3-F393-E0A9-E50E24DCCA9E")
	uartTXCharUUID  = ble.MustParse("6E400003-B5A3-F393-E0A9-E50E24DCCA9E")
)

var bleDevOnce sync.Once
var bleDevErr error

func ensureDevice() error {
	bleDevOnce.Do(func() {
		d, err := dev.NewDevice("default")
		if err != nil {
			bleDevErr = err
			return
		}
		ble.SetDefaultDevice(d)
	})
	return bleDevErr
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
	return b.connected && b.cln != nil
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
	if b.connected && b.cln != nil {
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
	ctx, cancel := context.WithTimeout(ctx, scanTimeout)
	defer cancel()

	var found ble.Advertisement
	foundCh := make(chan ble.Advertisement, 1)

	filter := func(a ble.Advertisement) bool {
		ok := false
		for _, s := range a.Services() {
			if s.Equal(uartServiceUUID) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
		if b.targetName == "" && b.targetAddr == "" {
			return strings.HasPrefix(a.LocalName(), "RNode ")
		}
		if b.targetAddr != "" && !strings.EqualFold(a.Addr().String(), b.targetAddr) {
			return false
		}
		if b.targetName != "" && a.LocalName() != b.targetName {
			return false
		}
		return true
	}

	go func() {
		_ = ble.Scan(ctx, false, func(a ble.Advertisement) {
			if filter(a) {
				select {
				case foundCh <- a:
				default:
				}
			}
		}, nil)
	}()

	select {
	case found = <-foundCh:
	case <-ctx.Done():
		return errors.New("ble scan timeout")
	}

	paired, err := bluezIsDeviceBonded(found.Addr().String())
	if err != nil {
		return err
	}
	if !paired {
		return errors.New("ble device is not bonded/paired")
	}

	connCtx, connCancel := context.WithTimeout(context.Background(), connectTimeout)
	defer connCancel()
	cln, err := ble.Dial(connCtx, found.Addr())
	if err != nil {
		return err
	}
	_, _ = cln.ExchangeMTU(512)

	prof, err := cln.DiscoverProfile(true)
	if err != nil {
		_ = cln.CancelConnection()
		return err
	}

	var rxChr, txChr *ble.Characteristic
	for _, s := range prof.Services {
		if !s.UUID.Equal(uartServiceUUID) {
			continue
		}
		for _, c := range s.Characteristics {
			switch {
			case c.UUID.Equal(uartRXCharUUID):
				rxChr = c
			case c.UUID.Equal(uartTXCharUUID):
				txChr = c
			}
		}
	}
	if rxChr == nil || txChr == nil {
		_ = cln.CancelConnection()
		return errors.New("ble uart characteristics not found")
	}

	b.mu.Lock()
	b.cln = cln
	b.rxChr = rxChr
	b.txChr = txChr
	b.connected = true
	b.disappeared = false
	b.mu.Unlock()

	go func() {
		<-cln.Disconnected()
		b.mu.Lock()
		b.connected = false
		b.disappeared = true
		b.mu.Unlock()
		select { case b.rxCh <- struct{}{}: default: }
		select { case b.reconnectCh <- struct{}{}: default: }
	}()

	if err := cln.Subscribe(txChr, false, func(req []byte) {
		b.rxMu.Lock()
		b.rxBuf.Write(req)
		b.rxMu.Unlock()
		select { case b.rxCh <- struct{}{}: default: }
	}); err != nil {
		_ = cln.CancelConnection()
		b.mu.Lock()
		b.connected = false
		b.cln = nil
		b.mu.Unlock()
		return err
	}

	select { case b.reconnectCh <- struct{}{}: default: }
	return nil
}

func (b *transport) Close() error {
	b.mu.Lock()
	cln := b.cln
	b.connected = false
	b.cln = nil
	b.rxChr = nil
	b.txChr = nil
	b.mu.Unlock()

	select { case <-b.stopCh: default: close(b.stopCh) }
	if b.txCond != nil {
		b.txCond.Broadcast()
	}
	if cln != nil {
		_ = cln.CancelConnection()
	}
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
	cln := b.cln
	rxChr := b.rxChr
	connected := b.connected
	b.mu.Unlock()

	if !connected || cln == nil || rxChr == nil {
		b.txMu.Lock()
		b.txQueue = append(b.txQueue, p...)
		if b.txCond != nil {
			b.txCond.Signal()
		}
		b.txMu.Unlock()
		return len(p), nil
	}

	for _, chunk := range ChunkByMaxWriteNoRsp(p, maxWriteNoRsp(cln)) {
		_ = cln.WriteCharacteristic(rxChr, chunk, false)
	}
	return len(p), nil
}

func (b *transport) reconnectLoop() {
	for {
		select {
		case <-b.stopCh:
			return
		case <-b.reconnectCh:
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

func (b *transport) txLoop() {
	for {
		select {
		case <-b.stopCh:
			return
		default:
		}
		b.mu.Lock()
		cln := b.cln
	rxChr := b.rxChr
		connected := b.connected
		b.mu.Unlock()
		if !connected || cln == nil || rxChr == nil {
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

		for _, part := range ChunkByMaxWriteNoRsp(buf, maxWriteNoRsp(cln)) {
			_ = cln.WriteCharacteristic(rxChr, part, false)
		}
	}
}

func maxWriteNoRsp(cln ble.Client) int {
	if cln == nil || cln.Conn() == nil {
		return 20
	}
	mtu := cln.Conn().TxMTU()
	if mtu <= 3 {
		return 20
	}
	n := mtu - 3
	if n < 20 {
		return 20
	}
	return n
}
