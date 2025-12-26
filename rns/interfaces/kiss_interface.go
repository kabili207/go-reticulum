package interfaces

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tarm/serial"
)

// Minimal KISS (no AX.25 header) driver. Frames are passed through as-is.

type kissConfig struct {
	Name         string
	Port         string
	Speed        int
	DataBits     int
	StopBits     serial.StopBits
	Parity       serial.Parity
	HWMTU        int
	BitrateGuess int
	ReadTimeout  time.Duration
	Preamble     int
	TxTail       int
	Persistence  int
	SlotTime     int
	FlowControl  bool
	FlowTimeout  time.Duration
	BeaconEvery  time.Duration
	BeaconData   []byte
}

type KISSDriver struct {
	iface *Interface
	cfg   kissConfig

	serialMu sync.Mutex
	serial   tarmSerialPort

	sendMu       sync.Mutex
	stopCh       chan struct{}
	running      atomic.Bool
	reconnecting atomic.Bool

	packetQueue    [][]byte
	interfaceReady bool
	flowLocked     time.Time
	firstTx        time.Time
}

func NewKISSInterface(name string, kv map[string]string) (*Interface, error) {
	cfg, err := parseKISSConfig(name, kv)
	if err != nil {
		return nil, err
	}

	iface := &Interface{
		Name:              name,
		Type:              "KISSInterface",
		IN:                true,
		OUT:               true,
		DriverImplemented: true,
		Bitrate:           cfg.BitrateGuess,
		HWMTU:             cfg.HWMTU,
	}
	// Match Python DEFAULT_IFAC_SIZE.
	if iface.IFACSize == 0 {
		iface.IFACSize = 8
	}

	driver := &KISSDriver{
		iface:  iface,
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
	iface.kiss = driver

	if err := driver.start(); err != nil {
		return iface, err
	}
	return iface, nil
}

func parseKISSConfig(name string, kv map[string]string) (kissConfig, error) {
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

	cfg := kissConfig{
		Name:         name,
		Port:         get("port"),
		Speed:        parseIntOr(get("speed"), 9600),
		DataBits:     parseIntOr(get("databits"), 8),
		StopBits:     serial.Stop1,
		Parity:       serial.ParityNone,
		HWMTU:        564,
		BitrateGuess: 1200,
		ReadTimeout:  time.Duration(parseIntOr(get("read_timeout"), 100)) * time.Millisecond,
		Preamble:     parseIntOr(get("preamble"), 350),
		TxTail:       parseIntOr(get("txtail"), 20),
		Persistence:  parseIntOr(get("persistence"), 64),
		SlotTime:     parseIntOr(get("slottime"), 20),
		FlowControl:  parseBoolOr(get("flow_control"), false),
		FlowTimeout:  time.Duration(parseIntOr(get("flow_control_timeout"), 5)) * time.Second,
	}
	if cfg.Port == "" {
		return cfg, fmt.Errorf("KISSInterface requires a serial port")
	}

	switch parseIntOr(get("stopbits"), 1) {
	case 2:
		cfg.StopBits = serial.Stop2
	default:
		cfg.StopBits = serial.Stop1
	}
	parity := strings.ToLower(get("parity"))
	switch parity {
	case "e", "even":
		cfg.Parity = serial.ParityEven
	case "o", "odd":
		cfg.Parity = serial.ParityOdd
	default:
		cfg.Parity = serial.ParityNone
	}

	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = 100 * time.Millisecond
	}

	if iv := parseIntOr(get("id_interval"), 0); iv > 0 {
		cfg.BeaconEvery = time.Duration(iv) * time.Second
	}
	if cs := get("id_callsign"); cs != "" {
		cfg.BeaconData = []byte(cs)
	}
	if cfg.FlowTimeout <= 0 {
		cfg.FlowTimeout = 5 * time.Second
	}

	return cfg, nil
}

func (d *KISSDriver) start() error {
	if DiagLogf != nil {
		DiagLogf(LogVerbose, "Opening serial port %s...", d.cfg.Port)
	}
	if err := d.openPort(); err != nil {
		if DiagLogf != nil {
			DiagLogf(LogError, "Could not open serial port %s: %v", d.cfg.Port, err)
		}
		return err
	}

	// Allow time for interface to initialise before config (Python uses 2s).
	kissSleep(2 * time.Second)

	if DiagLogf != nil {
		DiagLogf(LogInfo, "Serial port %s is now open", d.cfg.Port)
		DiagLogf(LogVerbose, "Configuring KISS interface parameters...")
	}
	if err := d.configureDevice(); err != nil {
		d.iface.Online = false
		d.closeSerial()
		if DiagLogf != nil {
			DiagLogf(LogError, "Could not configure KISS interface parameters: %v", err)
		}
		return err
	}
	d.interfaceReady = true
	if DiagLogf != nil {
		DiagLogf(LogInfo, "KISS interface configured")
	}

	go d.readLoop()
	return nil
}

func (d *KISSDriver) configureDevice() error {
	if err := d.sendCmd(kissCmdTxDelay, msToKISS(d.cfg.Preamble)); err != nil {
		return err
	}
	if err := d.sendCmd(kissCmdTxTail, msToKISS(d.cfg.TxTail)); err != nil {
		return err
	}
	if err := d.sendCmd(kissCmdP, byte(d.cfg.Persistence)); err != nil {
		return err
	}
	if err := d.sendCmd(kissCmdSlot, msToKISS(d.cfg.SlotTime)); err != nil {
		return err
	}
	// Python always attempts to enable READY notifications, regardless of whether
	// host-side flow control is enabled.
	if err := d.sendCmd(kissCmdReady, 0x01); err != nil {
		return err
	}
	return nil
}

func msToKISS(ms int) byte {
	if ms < 0 {
		ms = 0
	}
	v := ms / 10
	if v < 0 {
		v = 0
	}
	if v > 255 {
		v = 255
	}
	return byte(v)
}

func (d *KISSDriver) sendCmd(cmd byte, val byte) error {
	serial := d.getSerial()
	if serial == nil {
		return fmt.Errorf("serial not open")
	}
	frame := []byte{kissFEND, cmd, val, kissFEND}
	n, err := serial.Write(frame)
	if err != nil {
		return err
	}
	if n != len(frame) {
		return fmt.Errorf("short write for command 0x%02x", cmd)
	}
	return nil
}

func (d *KISSDriver) Close() {
	select {
	case <-d.stopCh:
	default:
		close(d.stopCh)
	}
	d.iface.Online = false
	d.closeSerial()
}

func (d *KISSDriver) setSerial(p tarmSerialPort) {
	d.serialMu.Lock()
	old := d.serial
	d.serial = p
	d.serialMu.Unlock()
	if old != nil && old != p {
		_ = old.Close()
	}
}

func (d *KISSDriver) getSerial() tarmSerialPort {
	d.serialMu.Lock()
	defer d.serialMu.Unlock()
	return d.serial
}

func (d *KISSDriver) openPort() error {
	c := &serial.Config{
		Name:        d.cfg.Port,
		Baud:        d.cfg.Speed,
		Size:        byte(d.cfg.DataBits),
		StopBits:    d.cfg.StopBits,
		Parity:      d.cfg.Parity,
		ReadTimeout: d.cfg.ReadTimeout,
	}
	port, err := openTarmSerialPort(c)
	if err != nil {
		return err
	}
	d.setSerial(port)
	d.iface.Online = true
	return nil
}

func (d *KISSDriver) closeSerial() {
	d.serialMu.Lock()
	if d.serial != nil {
		_ = d.serial.Close()
		d.serial = nil
	}
	d.serialMu.Unlock()
}

func (d *KISSDriver) ProcessOutgoing(payload []byte) {
	if len(payload) == 0 {
		return
	}
	d.sendMu.Lock()
	defer d.sendMu.Unlock()
	d.sendFrameLocked(payload, true)
}

func (d *KISSDriver) sendFrameLocked(payload []byte, startBeaconTimer bool) {
	if len(payload) == 0 {
		return
	}
	startBeaconTimer = startBeaconTimer && d.cfg.BeaconEvery > 0
	if len(d.cfg.BeaconData) > 0 && bytes.Equal(payload, d.cfg.BeaconData) {
		// Python: sending the ID payload resets first_tx and does not start beacon timer.
		d.firstTx = time.Time{}
		startBeaconTimer = false
	}
	frame := d.buildFrame(payload)

	serial := d.getSerial()
	if serial == nil {
		d.queueLocked(payload)
		d.startReconnect()
		return
	}
	if !d.interfaceReady {
		d.queueLocked(payload)
		return
	}

	if d.cfg.FlowControl {
		d.interfaceReady = false
		d.flowLocked = time.Now()
	}

	if n, err := serial.Write(frame); err != nil || n != len(frame) {
		if err == nil {
			err = fmt.Errorf("short write (%d/%d)", n, len(frame))
		}
		d.queueFrontLocked(payload)
		d.iface.Online = false
		if DiagLogf != nil {
			DiagLogf(LogError, "A serial port error occurred, the contained exception was: %v", err)
			DiagLogf(LogError, "The interface %s experienced an unrecoverable error and is now offline.", d.String())
			DiagLogf(LogError, "Reticulum will attempt to reconnect the interface periodically.")
		}
		d.closeSerial()
		d.startReconnect()
		return
	}

	atomic.AddUint64(&d.iface.TXB, uint64(len(payload)))
	if parent := d.iface.Parent; parent != nil {
		atomic.AddUint64(&parent.TXB, uint64(len(payload)))
	}
	if startBeaconTimer && d.firstTx.IsZero() {
		d.firstTx = time.Now()
	}
	if !d.cfg.FlowControl {
		d.interfaceReady = true
	}
}

func (d *KISSDriver) buildFrame(payload []byte) []byte {
	escaped := kissEscapeFrame(payload)
	frame := make([]byte, 0, len(escaped)+3)
	frame = append(frame, kissFEND, kissCmdData)
	frame = append(frame, escaped...)
	frame = append(frame, kissFEND)
	return frame
}

func (d *KISSDriver) readLoop() {
	defer d.startReconnect()
	d.running.Store(true)
	defer d.running.Store(false)

	var (
		inFrame    bool
		escape     bool
		command    byte = kissCmdUnknown
		dataBuffer      = make([]byte, 0, d.cfg.HWMTU)
		lastRead        = time.Now()
	)
	frameTimeout := d.cfg.ReadTimeout
	if frameTimeout <= 0 {
		frameTimeout = 100 * time.Millisecond
	}
	idleSleep := frameTimeout / 2
	if idleSleep <= 0 {
		idleSleep = 50 * time.Millisecond
	}

	for {
		select {
		case <-d.stopCh:
			return
		default:
		}
		serial := d.getSerial()
		if serial == nil {
			return
		}
		byteBuf := []byte{0}
		n, err := serial.Read(byteBuf)
		if err != nil {
			d.iface.Online = false
			if DiagLogf != nil {
				DiagLogf(LogError, "A serial port error occurred, the contained exception was: %v", err)
				DiagLogf(LogError, "The interface %s experienced an unrecoverable error and is now offline.", d.String())
				DiagLogf(LogError, "Reticulum will attempt to reconnect the interface periodically.")
			}
			d.closeSerial()
			return
		}
		if n == 0 {
			time.Sleep(idleSleep)
			if len(dataBuffer) > 0 && time.Since(lastRead) > frameTimeout {
				inFrame = false
				escape = false
				command = kissCmdUnknown
				dataBuffer = dataBuffer[:0]
			}
			if d.cfg.FlowControl && !d.interfaceReady && time.Since(d.flowLocked) > d.cfg.FlowTimeout {
				if DiagLogf != nil {
					DiagLogf(LogWarning, "Interface %s is unlocking flow control due to time-out. This should not happen. Your hardware might have missed a flow-control READY command, or maybe it does not support flow-control.", d.String())
				}
				d.processQueue()
			}
			if d.cfg.BeaconEvery > 0 && !d.firstTx.IsZero() {
				if time.Since(d.firstTx) > d.cfg.BeaconEvery {
					if DiagLogf != nil {
						DiagLogf(LogDebug, "Interface %s is transmitting beacon data: %s", d.String(), string(d.cfg.BeaconData))
					}
					d.sendMu.Lock()
					d.firstTx = time.Time{}
					frame := append([]byte(nil), d.cfg.BeaconData...)
					for len(frame) < 15 {
						frame = append(frame, 0x00)
					}
					d.sendFrameLocked(frame, false)
					d.sendMu.Unlock()
				}
			}
			continue
		}
		lastRead = time.Now()

		b := byteBuf[0]
		switch {
		case inFrame && b == kissFEND && command == kissCmdData:
			inFrame = false
			payload := append([]byte(nil), dataBuffer...)
			if InboundHandler != nil {
				InboundHandler(payload, d.iface)
			}
			atomic.AddUint64(&d.iface.RXB, uint64(len(payload)))
			if parent := d.iface.Parent; parent != nil {
				atomic.AddUint64(&parent.RXB, uint64(len(payload)))
			}
			dataBuffer = dataBuffer[:0]
		case b == kissFEND:
			inFrame = true
			command = kissCmdUnknown
			dataBuffer = dataBuffer[:0]
			escape = false
		case inFrame && len(dataBuffer) < d.cfg.HWMTU:
			if len(dataBuffer) == 0 && command == kissCmdUnknown {
				// Strip port nibble; we only support one HDLC port for now.
				command = b & 0x0F
			} else if command == kissCmdData {
				if b == kissFESC {
					escape = true
				} else {
					if escape {
						if b == kissTFEND {
							b = kissFEND
						} else if b == kissTFESC {
							b = kissFESC
						}
						escape = false
					}
					dataBuffer = append(dataBuffer, b)
				}
			} else if command == kissCmdReady {
				d.processQueue()
			}
		}
	}
}

func (d *KISSDriver) processQueue() {
	d.sendMu.Lock()
	defer d.sendMu.Unlock()
	if len(d.packetQueue) == 0 {
		d.interfaceReady = true
		return
	}
	next := d.packetQueue[0]
	d.packetQueue = d.packetQueue[1:]
	d.interfaceReady = true
	d.sendFrameLocked(next, true)
}

func (d *KISSDriver) queueLocked(data []byte) {
	d.packetQueue = append(d.packetQueue, append([]byte(nil), data...))
}

func (d *KISSDriver) queueFrontLocked(data []byte) {
	d.packetQueue = append([][]byte{append([]byte(nil), data...)}, d.packetQueue...)
}

func (d *KISSDriver) startReconnect() {
	if !d.reconnecting.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer d.reconnecting.Store(false)
		for {
			select {
			case <-d.stopCh:
				return
			case <-time.After(5 * time.Second):
			}
			if DiagLogf != nil {
				DiagLogf(LogVerbose, "Attempting to reconnect serial port %s for %s...", d.cfg.Port, d.String())
			}
			if err := d.openPort(); err != nil {
				if DiagLogf != nil {
					DiagLogf(LogError, "Error while reconnecting port, the contained exception was: %v", err)
				}
				continue
			}
			kissSleep(2 * time.Second)
			d.sendMu.Lock()
			if err := d.configureDevice(); err != nil {
				d.sendMu.Unlock()
				d.iface.Online = false
				d.closeSerial()
				if DiagLogf != nil {
					DiagLogf(LogError, "Could not configure KISS interface parameters: %v", err)
				}
				continue
			}
			d.interfaceReady = true
			d.sendMu.Unlock()
			if DiagLogf != nil {
				DiagLogf(LogInfo, "Reconnected serial port for %s", d.String())
			}
			go d.readLoop()
			return
		}
	}()
}

func (d *KISSDriver) String() string {
	if d == nil || d.cfg.Name == "" {
		return "KISSInterface"
	}
	return "KISSInterface[" + d.cfg.Name + "]"
}
