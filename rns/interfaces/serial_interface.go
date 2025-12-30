package interfaces

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	serial "go.bug.st/serial"
)

type serialConfig struct {
	Name     string
	Port     string
	Speed    int
	DataBits int
	Parity   serial.Parity
	StopBits serial.StopBits
	Timeout  time.Duration
	HWMTU    int
}

type SerialDriver struct {
	iface *Interface
	cfg   serialConfig

	serialMu sync.Mutex
	port     serial.Port

	sendMu sync.Mutex

	stopCh       chan struct{}
	stopped      atomic.Bool
	reconnecting atomic.Bool
}

func NewSerialInterface(name string, kv map[string]string) (*Interface, error) {
	cfg, err := parseSerialConfig(strings.TrimSpace(name), kv)
	if err != nil {
		return nil, err
	}

	iface := &Interface{
		Name:              cfg.Name,
		Type:              "SerialInterface",
		IN:                true,
		OUT:               true,
		DriverImplemented: true,
		Bitrate:           cfg.Speed, // Python: self.bitrate = self.speed
		HWMTU:             cfg.HWMTU,
	}
	// Match Python SerialInterface.DEFAULT_IFAC_SIZE.
	if iface.IFACSize == 0 {
		iface.IFACSize = 8
	}

	driver := &SerialDriver{
		iface:  iface,
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
	iface.serial = driver

	if err := driver.start(); err != nil {
		return iface, err
	}
	return iface, nil
}

func parseSerialConfig(name string, kv map[string]string) (serialConfig, error) {
	get := func(key string) string {
		if kv == nil {
			return ""
		}
		if v, ok := kv[key]; ok {
			return strings.TrimSpace(v)
		}
		lower := strings.ToLower(key)
		for k, v := range kv {
			if strings.ToLower(k) == lower {
				return strings.TrimSpace(v)
			}
		}
		return ""
	}

	cfg := serialConfig{
		Name:     name,
		Port:     get("port"),
		Speed:    parseIntOr(get("speed"), 9600),
		DataBits: parseIntOr(get("databits"), 8),
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
		Timeout:  time.Duration(parseIntOr(get("timeout"), 100)) * time.Millisecond, // Python: self.timeout=100
		HWMTU:    564,                                                               // Python: HW_MTU
	}
	if cfg.Port == "" {
		return cfg, errors.New("no port specified for serial interface")
	}
	if cfg.Speed <= 0 {
		cfg.Speed = 9600
	}
	if cfg.DataBits == 0 {
		cfg.DataBits = 8
	}

	switch parseIntOr(get("stopbits"), 1) {
	case 2:
		cfg.StopBits = serial.TwoStopBits
	default:
		cfg.StopBits = serial.OneStopBit
	}

	switch strings.ToLower(get("parity")) {
	case "e", "even":
		cfg.Parity = serial.EvenParity
	case "o", "odd":
		cfg.Parity = serial.OddParity
	case "n", "none", "":
		cfg.Parity = serial.NoParity
	default:
		cfg.Parity = serial.NoParity
	}

	if cfg.Timeout <= 0 {
		cfg.Timeout = 100 * time.Millisecond
	}

	return cfg, nil
}

func (d *SerialDriver) start() error {
	if DiagLogf != nil {
		DiagLogf(LogVerbose, "Opening serial port %s...", d.cfg.Port)
	}
	if err := d.openPort(); err != nil {
		if DiagLogf != nil {
			DiagLogf(LogError, "Could not open serial port %s: %v", d.cfg.Port, err)
		}
		return err
	}

	// Python: sleep(0.5) before starting reader.
	time.Sleep(500 * time.Millisecond)
	d.iface.Online = true
	if DiagLogf != nil {
		DiagLogf(LogVerbose, "Serial port %s is now open", d.cfg.Port)
	}

	go d.readLoop()
	return nil
}

func (d *SerialDriver) Close() {
	if d == nil {
		return
	}
	if d.stopped.Swap(true) {
		return
	}
	select {
	case <-d.stopCh:
	default:
		close(d.stopCh)
	}
	d.iface.Online = false
	d.closePort()
}

func (d *SerialDriver) openPort() error {
	mode := &serial.Mode{
		BaudRate: d.cfg.Speed,
		DataBits: d.cfg.DataBits,
		Parity:   d.cfg.Parity,
		StopBits: d.cfg.StopBits,
	}
	p, err := serial.Open(d.cfg.Port, mode)
	if err != nil {
		return err
	}

	// Python: timeout=0 (non-blocking). We mimic with 0 read timeout, making
	// Read return (0,nil) immediately when no data is available.
	if err := p.SetReadTimeout(0); err != nil {
		_ = p.Close()
		return err
	}

	d.serialMu.Lock()
	old := d.port
	d.port = p
	d.serialMu.Unlock()
	if old != nil && old != p {
		_ = old.Close()
	}
	return nil
}

func (d *SerialDriver) closePort() {
	d.serialMu.Lock()
	p := d.port
	d.port = nil
	d.serialMu.Unlock()
	if p != nil {
		_ = p.Close()
	}
}

func (d *SerialDriver) getPort() serial.Port {
	d.serialMu.Lock()
	p := d.port
	d.serialMu.Unlock()
	return p
}

func (d *SerialDriver) ProcessOutgoing(data []byte) {
	if d == nil || len(data) == 0 || d.stopped.Load() {
		return
	}
	if !d.iface.Online {
		return
	}

	framed := append([]byte{hdlcFlag}, hdlcEscape(data)...)
	framed = append(framed, hdlcFlag)

	p := d.getPort()
	if p == nil {
		return
	}

	d.sendMu.Lock()
	n, err := p.Write(framed)
	d.sendMu.Unlock()

	if err != nil || n != len(framed) {
		if DiagLogf != nil {
			DiagLogf(LogError, "Exception occurred while transmitting via %s", d.iface)
			if err != nil {
				DiagLogf(LogError, "The contained exception was: %v", err)
			} else {
				DiagLogf(LogError, "The contained exception was: short write (%d/%d)", n, len(framed))
			}
		}
		d.iface.Online = false
		d.closePort()
		go d.reconnectLoop()
		return
	}

	atomic.AddUint64(&d.iface.TXB, uint64(n))
	if parent := d.iface.Parent; parent != nil {
		atomic.AddUint64(&parent.TXB, uint64(n))
	}
}

func (d *SerialDriver) readLoop() {
	defer func() {
		d.iface.Online = false
		d.closePort()
		if !d.stopped.Load() {
			d.reconnectLoop()
		}
	}()

	var (
		inFrame    = false
		escape     = false
		dataBuf    bytes.Buffer
		lastReadMS = time.Now().UnixMilli()
		timeoutMS  = d.cfg.Timeout.Milliseconds()
	)

	readBuf := make([]byte, 4096)

	for !d.stopped.Load() {
		select {
		case <-d.stopCh:
			return
		default:
		}

		p := d.getPort()
		if p == nil {
			return
		}

		n, err := p.Read(readBuf)
		if err != nil {
			d.handlePortError(err)
			return
		}

		if n == 0 {
			// "no data": mirror Python behaviour (clear partial frame after timeout).
			timeSince := time.Now().UnixMilli() - lastReadMS
			if dataBuf.Len() > 0 && timeSince > timeoutMS {
				dataBuf.Reset()
				inFrame = false
				escape = false
			}
			time.Sleep(80 * time.Millisecond) // Python: sleep(0.08)
			continue
		}

		lastReadMS = time.Now().UnixMilli()

		for i := 0; i < n; i++ {
			b := readBuf[i]

			if inFrame && b == hdlcFlag {
				inFrame = false
				frame := append([]byte(nil), dataBuf.Bytes()...)
				dataBuf.Reset()
				if len(frame) > 0 {
					atomic.AddUint64(&d.iface.RXB, uint64(len(frame)))
					if parent := d.iface.Parent; parent != nil {
						atomic.AddUint64(&parent.RXB, uint64(len(frame)))
					}
					if InboundHandler != nil {
						InboundHandler(frame, d.iface)
					}
				}
				continue
			}
			if b == hdlcFlag {
				inFrame = true
				escape = false
				dataBuf.Reset()
				continue
			}

			if inFrame && dataBuf.Len() < d.cfg.HWMTU {
				if b == hdlcEsc {
					escape = true
					continue
				}
				if escape {
					if b == (hdlcFlag ^ hdlcEscMask) {
						b = hdlcFlag
					} else if b == (hdlcEsc ^ hdlcEscMask) {
						b = hdlcEsc
					}
					escape = false
				}
				_ = dataBuf.WriteByte(b)
			}
		}
	}
}

func (d *SerialDriver) reconnectLoop() {
	if d == nil || d.stopped.Load() {
		return
	}
	if !d.reconnecting.CompareAndSwap(false, true) {
		return
	}
	defer d.reconnecting.Store(false)

	for !d.stopped.Load() {
		time.Sleep(5 * time.Second)

		if DiagLogf != nil {
			DiagLogf(LogVerbose, "Attempting to reconnect serial port %s for %s...", d.cfg.Port, d.iface)
		}
		if err := d.openPort(); err != nil {
			if DiagLogf != nil {
				DiagLogf(LogError, "Error while reconnecting port, the contained exception was: %v", err)
			}
			continue
		}

		time.Sleep(500 * time.Millisecond)
		d.iface.Online = true
		if DiagLogf != nil {
			DiagLogf(LogVerbose, "Reconnected serial port for %s", d.iface)
		}

		go d.readLoop()
		return
	}
}

func (d *SerialDriver) handlePortError(err error) {
	d.iface.Online = false
	if DiagLogf != nil {
		DiagLogf(LogError, "A serial port error occurred, the contained exception was: %v", err)
		DiagLogf(LogError, "The interface %s experienced an unrecoverable error and is now offline.", d.iface)
	}

	if PanicOnInterfaceErrorProvider != nil && PanicOnInterfaceErrorProvider() {
		if PanicFunc != nil {
			PanicFunc()
			return
		}
		panic(fmt.Errorf("serial interface error: %w", err))
	}

	if DiagLogf != nil {
		DiagLogf(LogError, "Reticulum will attempt to reconnect the interface periodically.")
	}
}
