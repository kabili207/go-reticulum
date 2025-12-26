package interfaces

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tarm/serial"
)

const (
	kissFEND       = 0xC0
	kissFESC       = 0xDB
	kissTFEND      = 0xDC
	kissTFESC      = 0xDD
	kissCmdUnknown = 0xFE
	kissCmdData    = 0x00
	kissCmdTxDelay = 0x01
	kissCmdP       = 0x02
	kissCmdSlot    = 0x03
	kissCmdTxTail  = 0x04
	kissCmdReady   = 0x0F

	ax25CtrlUI     = 0x03
	ax25PIDNoLayer = 0xF0
	ax25HeaderSize = 16
)

type ax25kissConfig struct {
	Name         string
	Port         string
	Speed        int
	DataBits     int
	StopBits     serial.StopBits
	Parity       serial.Parity
	Callsign     []byte
	SSID         int
	DstCall      []byte
	DstSSID      int
	Preamble     int
	TxTail       int
	Persistence  int
	SlotTime     int
	FlowControl  bool
	FlowTimeout  time.Duration
	ReadTimeout  time.Duration
	FrameTimeout time.Duration
	HWMTU        int
	BitrateGuess int
}

type AX25KISSDriver struct {
	iface *Interface
	cfg   ax25kissConfig

	serialMu sync.Mutex
	serial   tarmSerialPort

	sendMu         sync.Mutex
	packetQueue    [][]byte
	interfaceReady bool
	flowLocked     time.Time

	stopCh       chan struct{}
	running      atomic.Bool
	reconnecting atomic.Bool

	addrHeader []byte
}

// NewAX25KISSInterface creates and starts an AX.25 KISS driver.
func NewAX25KISSInterface(name string, kv map[string]string) (*Interface, error) {
	cfg, err := parseAX25KISSConfig(name, kv)
	if err != nil {
		return nil, err
	}

	iface := &Interface{
		Name:              name,
		Type:              "AX25KISSInterface",
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

	driver := &AX25KISSDriver{
		iface:      iface,
		cfg:        cfg,
		stopCh:     make(chan struct{}),
		addrHeader: buildAX25Address(cfg),
	}
	if len(driver.addrHeader) != ax25HeaderSize-2 {
		return nil, errors.New("invalid AX.25 address header")
	}
	iface.ax25 = driver

	if err := driver.start(); err != nil {
		return nil, err
	}
	return iface, nil
}

func parseAX25KISSConfig(name string, kv map[string]string) (ax25kissConfig, error) {
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

	cfg := ax25kissConfig{
		Name:         name,
		Port:         get("port"),
		Speed:        parseIntOr(get("speed"), 9600),
		DataBits:     parseIntOr(get("databits"), 8),
		StopBits:     serial.Stop1,
		Parity:       serial.ParityNone,
		SSID:         parseIntOr(get("ssid"), -1),
		DstCall:      []byte("APZRNS"),
		DstSSID:      0,
		Preamble:     parseIntOr(get("preamble"), 350),
		TxTail:       parseIntOr(get("txtail"), 20),
		Persistence:  parseIntOr(get("persistence"), 64),
		SlotTime:     parseIntOr(get("slottime"), 20),
		FlowControl:  parseBoolOr(get("flow_control"), false),
		FlowTimeout:  time.Duration(parseIntOr(get("flow_control_timeout"), 5)) * time.Second,
		ReadTimeout:  time.Duration(parseIntOr(get("read_timeout"), 100)) * time.Millisecond,
		FrameTimeout: time.Duration(parseIntOr(get("timeout"), 100)) * time.Millisecond,
		HWMTU:        564,
		BitrateGuess: 1200,
	}

	if cfg.Port == "" {
		return cfg, errors.New("AX25KISSInterface requires a serial port")
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

	switch parseIntOr(get("stopbits"), 1) {
	case 2:
		cfg.StopBits = serial.Stop2
	default:
		cfg.StopBits = serial.Stop1
	}

	callsign := strings.ToUpper(get("callsign"))
	if len(callsign) < 3 || len(callsign) > 6 {
		return cfg, fmt.Errorf("invalid callsign for AX25KISSInterface %q", name)
	}
	cfg.Callsign = []byte(callsign)
	if cfg.SSID < 0 || cfg.SSID > 15 {
		return cfg, fmt.Errorf("invalid SSID for AX25KISSInterface %q", name)
	}

	if dst := strings.ToUpper(get("dst_call")); dst != "" {
		if len(dst) < 3 || len(dst) > 6 {
			return cfg, fmt.Errorf("invalid destination callsign for AX25KISSInterface %q", name)
		}
		cfg.DstCall = []byte(dst)
	}
	if v := get("dst_ssid"); v != "" {
		if parsed := parseIntOr(v, cfg.DstSSID); parsed >= 0 && parsed <= 15 {
			cfg.DstSSID = parsed
		}
	}

	if cfg.FlowTimeout <= 0 {
		cfg.FlowTimeout = 5 * time.Second
	}
	if cfg.FrameTimeout <= 0 {
		cfg.FrameTimeout = 100 * time.Millisecond
	}

	return cfg, nil
}

func buildAX25Address(cfg ax25kissConfig) []byte {
	addr := make([]byte, 0, 14)
	for i := 0; i < 6; i++ {
		if i < len(cfg.DstCall) {
			addr = append(addr, cfg.DstCall[i]<<1)
		} else {
			addr = append(addr, 0x20)
		}
	}
	addr = append(addr, 0x60|(byte(cfg.DstSSID)<<1))

	for i := 0; i < 6; i++ {
		if i < len(cfg.Callsign) {
			addr = append(addr, cfg.Callsign[i]<<1)
		} else {
			addr = append(addr, 0x20)
		}
	}
	addr = append(addr, 0x61|(byte(cfg.SSID)<<1))
	return addr
}

func (d *AX25KISSDriver) start() error {
	if err := d.openPort(); err != nil {
		return err
	}
	if err := d.configureDevice(); err != nil {
		d.closeSerial()
		return err
	}
	go d.readLoop()
	return nil
}

func (d *AX25KISSDriver) openPort() error {
	if DiagLogf != nil {
		DiagLogf(LogVerbose, "Opening serial port %s...", d.cfg.Port)
	}
	cfg := &serial.Config{
		Name:        d.cfg.Port,
		Baud:        d.cfg.Speed,
		Size:        byte(d.cfg.DataBits),
		Parity:      d.cfg.Parity,
		StopBits:    d.cfg.StopBits,
		ReadTimeout: d.cfg.ReadTimeout,
	}
	sp, err := openTarmSerialPort(cfg)
	if err != nil {
		return err
	}
	d.serialMu.Lock()
	d.serial = sp
	d.serialMu.Unlock()
	return nil
}

func (d *AX25KISSDriver) configureDevice() error {
	ax25Sleep(2 * time.Second)
	if DiagLogf != nil {
		DiagLogf(LogInfo, "Serial port %s is now open", d.cfg.Port)
		DiagLogf(LogInfo, "Configuring AX.25 KISS interface parameters...")
	}
	if err := d.setPreamble(d.cfg.Preamble); err != nil {
		return err
	}
	if err := d.setTxTail(d.cfg.TxTail); err != nil {
		return err
	}
	if err := d.setPersistence(d.cfg.Persistence); err != nil {
		return err
	}
	if err := d.setSlotTime(d.cfg.SlotTime); err != nil {
		return err
	}
	if err := d.setFlowControl(); err != nil {
		return err
	}

	d.sendMu.Lock()
	d.interfaceReady = true
	d.sendMu.Unlock()

	d.iface.Online = true
	if DiagLogf != nil {
		DiagLogf(LogInfo, "AX.25 KISS interface configured")
	}
	return nil
}

func (d *AX25KISSDriver) Close() {
	select {
	case <-d.stopCh:
	default:
		close(d.stopCh)
	}
	d.serialMu.Lock()
	if d.serial != nil {
		_ = d.serial.Close()
		d.serial = nil
	}
	d.serialMu.Unlock()
	d.iface.Online = false
}

func (d *AX25KISSDriver) ProcessOutgoing(data []byte) {
	if len(data) == 0 {
		return
	}
	if !d.iface.Online {
		return
	}
	payload := append([]byte(nil), data...)
	if len(payload) > d.cfg.HWMTU {
		payload = payload[:d.cfg.HWMTU]
	}

	d.sendMu.Lock()
	defer d.sendMu.Unlock()
	if d.interfaceReady {
		d.sendFrameLocked(payload)
	} else {
		d.packetQueue = append(d.packetQueue, payload)
	}
}

func (d *AX25KISSDriver) sendFrameLocked(payload []byte) {
	frame := d.buildFrame(payload)
	if frame == nil {
		return
	}
	serial := d.getSerial()
	if serial == nil {
		d.packetQueue = append([][]byte{payload}, d.packetQueue...)
		d.interfaceReady = true
		return
	}

	if d.cfg.FlowControl {
		d.interfaceReady = false
		d.flowLocked = time.Now()
	}

	written, err := serial.Write(frame)
	if err != nil || written != len(frame) {
		d.packetQueue = append([][]byte{payload}, d.packetQueue...)
		d.interfaceReady = true
		if err == nil {
			err = fmt.Errorf("AX.25 interface only wrote %d bytes of %d", written, len(frame))
		}
		go d.handleSerialError(err)
		return
	}

	atomic.AddUint64(&d.iface.TXB, uint64(len(payload)))
	if !d.cfg.FlowControl {
		d.interfaceReady = true
	}
}

func (d *AX25KISSDriver) buildFrame(payload []byte) []byte {
	buf := make([]byte, 0, len(d.addrHeader)+2+len(payload)+4)
	buf = append(buf, d.addrHeader...)
	buf = append(buf, ax25CtrlUI, ax25PIDNoLayer)
	buf = append(buf, payload...)
	escaped := kissEscapeFrame(buf)
	frame := make([]byte, 0, len(escaped)+3)
	frame = append(frame, kissFEND, kissCmdData)
	frame = append(frame, escaped...)
	frame = append(frame, kissFEND)
	return frame
}

func kissEscapeFrame(data []byte) []byte {
	out := make([]byte, 0, len(data))
	for _, b := range data {
		switch b {
		case kissFESC:
			out = append(out, kissFESC, kissTFESC)
		case kissFEND:
			out = append(out, kissFESC, kissTFEND)
		default:
			out = append(out, b)
		}
	}
	return out
}

func (d *AX25KISSDriver) processQueue() {
	d.sendMu.Lock()
	defer d.sendMu.Unlock()
	d.processQueueLocked()
}

func (d *AX25KISSDriver) processQueueLocked() {
	if len(d.packetQueue) == 0 {
		d.interfaceReady = true
		return
	}
	next := d.packetQueue[0]
	d.packetQueue = d.packetQueue[1:]
	d.interfaceReady = true
	d.sendFrameLocked(next)
}

func (d *AX25KISSDriver) readLoop() {
	defer d.startReconnect()
	d.running.Store(true)
	defer d.running.Store(false)

	var (
		inFrame    bool
		escape     bool
		command    byte = kissCmdUnknown
		dataBuffer      = make([]byte, 0, d.cfg.HWMTU+ax25HeaderSize)
		lastRead        = time.Now()
	)
	frameTimeout := d.cfg.FrameTimeout
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
			d.handleSerialError(err)
			return
		}
		if n == 0 {
			if len(dataBuffer) > 0 && time.Since(lastRead) > frameTimeout {
				dataBuffer = dataBuffer[:0]
				inFrame = false
				command = kissCmdUnknown
				escape = false
			}
			time.Sleep(idleSleep)
			if d.cfg.FlowControl {
				d.sendMu.Lock()
				if !d.interfaceReady && time.Since(d.flowLocked) > d.cfg.FlowTimeout {
					if DiagLogf != nil {
						DiagLogf(
							LogWarning,
							"Interface %s is unlocking flow control due to time-out. This should not happen. Your hardware might have missed a flow-control READY command, or maybe it does not support flow-control.",
							d.iface,
						)
					}
					d.processQueueLocked()
				}
				d.sendMu.Unlock()
			}
			continue
		}

		lastRead = time.Now()
		b := byteBuf[0]
		switch {
		case inFrame && b == kissFEND && command == kissCmdData:
			inFrame = false
			d.processIncoming(dataBuffer)
			dataBuffer = dataBuffer[:0]
		case b == kissFEND:
			inFrame = true
			command = kissCmdUnknown
			dataBuffer = dataBuffer[:0]
			escape = false
		case inFrame && len(dataBuffer) < d.cfg.HWMTU+ax25HeaderSize:
			if len(dataBuffer) == 0 && command == kissCmdUnknown {
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

func (d *AX25KISSDriver) processIncoming(data []byte) {
	if len(data) <= ax25HeaderSize {
		return
	}
	payload := append([]byte(nil), data[ax25HeaderSize:]...)
	// Python increments RXB by full received AX.25 frame length (including header),
	// but forwards only the payload to the transport.
	atomic.AddUint64(&d.iface.RXB, uint64(len(data)))
	if InboundHandler != nil {
		InboundHandler(payload, d.iface)
	}
}

func (d *AX25KISSDriver) getSerial() tarmSerialPort {
	d.serialMu.Lock()
	defer d.serialMu.Unlock()
	return d.serial
}

func (d *AX25KISSDriver) handleSerialError(err error) {
	d.iface.Online = false
	if DiagLogf != nil {
		DiagLogf(LogError, "A serial port error occurred, the contained exception was: %v", err)
		DiagLogf(LogError, "The interface %s experienced an unrecoverable error and is now offline.", d.iface)
		DiagLogf(LogError, "Reticulum will attempt to reconnect the interface periodically.")
	}
	d.serialMu.Lock()
	if d.serial != nil {
		_ = d.serial.Close()
		d.serial = nil
	}
	d.serialMu.Unlock()
}

func (d *AX25KISSDriver) startReconnect() {
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
				DiagLogf(LogVerbose, "Attempting to reconnect serial port %s for %s...", d.cfg.Port, d.iface)
			}
			if err := d.openPort(); err != nil {
				if DiagLogf != nil {
					DiagLogf(LogError, "Error while reconnecting port, the contained exception was: %v", err)
				}
				continue
			}
			if err := d.configureDevice(); err != nil {
				if DiagLogf != nil {
					DiagLogf(LogError, "Error while reconnecting port, the contained exception was: %v", err)
				}
				d.closeSerial()
				continue
			}
			if DiagLogf != nil {
				DiagLogf(LogInfo, "Reconnected serial port for %s", d.iface)
			}
			go d.readLoop()
			return
		}
	}()
}

func (d *AX25KISSDriver) closeSerial() {
	d.serialMu.Lock()
	if d.serial != nil {
		_ = d.serial.Close()
		d.serial = nil
	}
	d.serialMu.Unlock()
}

func (d *AX25KISSDriver) setPreamble(ms int) error {
	val := clampByte(ms / 10)
	return d.writeKISSCommand(kissCmdTxDelay, val)
}

func (d *AX25KISSDriver) setTxTail(ms int) error {
	val := clampByte(ms / 10)
	return d.writeKISSCommand(kissCmdTxTail, val)
}

func (d *AX25KISSDriver) setPersistence(p int) error {
	val := clampByte(p)
	return d.writeKISSCommand(kissCmdP, val)
}

func (d *AX25KISSDriver) setSlotTime(ms int) error {
	val := clampByte(ms / 10)
	return d.writeKISSCommand(kissCmdSlot, val)
}

func (d *AX25KISSDriver) setFlowControl() error {
	return d.writeKISSCommand(kissCmdReady, 0x01)
}

func (d *AX25KISSDriver) writeKISSCommand(cmd byte, value byte) error {
	frame := []byte{kissFEND, cmd, value, kissFEND}
	serial := d.getSerial()
	if serial == nil {
		return errors.New("serial port not open")
	}
	written, err := serial.Write(frame)
	if err != nil {
		return err
	}
	if written != len(frame) {
		return fmt.Errorf("short write for command 0x%02x", cmd)
	}
	return nil
}

func clampByte(v int) byte {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return byte(v)
}

func parseIntOr(value string, def int) int {
	if strings.TrimSpace(value) == "" {
		return def
	}
	var n int
	if _, err := fmt.Sscanf(value, "%d", &n); err == nil {
		return n
	}
	return def
}

func parseBoolOr(value string, def bool) bool {
	if value == "" {
		return def
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}
