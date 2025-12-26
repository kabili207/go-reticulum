package main

import (
	"bytes"
	"errors"
	"fmt"
	"crypto/md5"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

// simSerialPort is a minimal in-process RNode emulator that speaks the same KISS framing
// expected by cmd/rnodeconf. It is intended for deterministic tests and development
// without hardware.
//
// Usage: pass a port string like "sim://esp32" to rnodeconf (darwin/linux).
type simSerialPort struct {
	name string

	mu     sync.Mutex
	closed bool
	readQ  bytes.Buffer

	platform byte
	mcu      byte
	board    byte

	major byte
	minor byte

	deviceHash        []byte
	firmwareHash      []byte
	firmwareHashTarget []byte

	eeprom    []byte
	cfgSector []byte

	// behaviour toggles (parsed from sim:// URL)
	extractResult string // "ok" | "fail180" | "fail182"
	flashResult   string // "ok" | "fail1"
}

func newSimSerialPort(name string) (*simSerialPort, error) {
	p := &simSerialPort{name: name}

	// Defaults: ESP32 RNode-ish values.
	p.platform = ROM_PLATFORM_ESP32
	p.mcu = ROM_MCU_ESP32
	p.board = 0x31 // ROM.BOARD_RNODE in Python
	p.major = 2
	p.minor = 5

	p.deviceHash = bytes.Repeat([]byte{0x11}, 32)
	p.firmwareHashTarget = bytes.Repeat([]byte{0x22}, 32)
	p.firmwareHash = bytes.Repeat([]byte{0x33}, 32)

	p.eeprom = makeSimEEPROM()
	p.cfgSector = makeSimCfgSector()

	p.extractResult = "ok"
	p.flashResult = "ok"

	if err := p.applyURLParams(name); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *simSerialPort) applyURLParams(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return nil
	}
	q := u.Query()
	plat := q.Get("platform")
	if plat == "" {
		// Allow shorthand like "sim://avr" (host becomes the selector).
		plat = strings.ToLower(strings.TrimSpace(u.Host))
	}
	switch plat {
	case "", "esp32":
		p.platform = ROM_PLATFORM_ESP32
		p.mcu = ROM_MCU_ESP32
	case "avr":
		p.platform = ROM_PLATFORM_AVR
		p.mcu = ROM_MCU_1284P
	case "nrf52":
		p.platform = ROM_PLATFORM_NRF52
		p.mcu = ROM_MCU_NRF52
	}
	if v := q.Get("fw"); v != "" {
		// fw=2.5 -> major=2 minor=5
		if parts := bytes.Split([]byte(v), []byte(".")); len(parts) == 2 {
			ma, _ := strconv.Atoi(string(parts[0]))
			mi, _ := strconv.Atoi(string(parts[1]))
			if ma >= 0 && ma <= 255 {
				p.major = byte(ma)
			}
			if mi >= 0 && mi <= 255 {
				p.minor = byte(mi)
			}
		}
	}
	if q.Get("provisioned") == "0" {
		// break checksum so parseEEPROM yields Provisioned=false
		if len(p.eeprom) > romAddrChecksum {
			p.eeprom[romAddrChecksum] ^= 0xFF
		}
	}
	if v := q.Get("extract"); v != "" {
		p.extractResult = v
	}
	if v := q.Get("flash"); v != "" {
		p.flashResult = v
	}
	return nil
}

func makeSimEEPROM() []byte {
	// Minimal, provisioned-looking EEPROM buffer large enough for parseEEPROM().
	// Signature is not valid by default (tests can assert Unverified behaviour).
	eeprom := make([]byte, 256)
	eeprom[romAddrProduct] = 0x03 // ROM.PRODUCT_RNODE
	eeprom[romAddrModel] = 0xA1
	eeprom[romAddrHWRev] = 0x01
	copy(eeprom[romAddrSerial:romAddrSerial+4], []byte{0, 0, 0, 1})
	copy(eeprom[romAddrMade:romAddrMade+4], []byte{0, 0, 0, 2})
	// Compute a correct checksum so the device is provisioned, but keep signature invalid.
	checksummedInfo := make([]byte, 0, 11)
	checksummedInfo = append(checksummedInfo, eeprom[romAddrProduct], eeprom[romAddrModel], eeprom[romAddrHWRev])
	checksummedInfo = append(checksummedInfo, eeprom[romAddrSerial:romAddrSerial+4]...)
	checksummedInfo = append(checksummedInfo, eeprom[romAddrMade:romAddrMade+4]...)
	sum := md5.Sum(checksummedInfo)
	copy(eeprom[romAddrChecksum:romAddrChecksum+16], sum[:])
	eeprom[romAddrInfoLock] = romInfoLockByte

	// Mark some config sections as present for nicer `--config` output.
	for _, addr := range []int{
		romAddrConfBT, romAddrConfWiFi, romAddrConfDInt, romAddrConfPSet, romAddrConfBSet,
	} {
		if addr >= 0 && addr < len(eeprom) {
			eeprom[addr] = confOKByte
		}
	}
	return eeprom
}

func makeSimCfgSector() []byte {
	// Minimal config sector.
	return make([]byte, 256)
}

func (p *simSerialPort) Name() string { return p.name }
func (p *simSerialPort) IsOpen() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return !p.closed
}
func (p *simSerialPort) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}
func (p *simSerialPort) BytesAvailable() (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.readQ.Len(), nil
}

func (p *simSerialPort) Read(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return 0, errors.New("closed")
	}
	return p.readQ.Read(b)
}

func (p *simSerialPort) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return 0, errors.New("closed")
	}
	// Parse concatenated KISS frames: FEND <cmd> <data...> FEND
	i := 0
	for i < len(b) {
		if b[i] != KISS_FEND {
			i++
			continue
		}
		// find next FEND
		j := i + 1
		for j < len(b) && b[j] != KISS_FEND {
			j++
		}
		if j >= len(b) {
			break
		}
		frame := b[i+1 : j]
		if len(frame) >= 1 {
			cmd := frame[0]
			data := frame[1:]
			p.handleFrame(cmd, data)
		}
		i = j + 1
	}
	return len(b), nil
}

func (p *simSerialPort) SetDTR(bool) error { return nil }
func (p *simSerialPort) SetRTS(bool) error { return nil }

func (p *simSerialPort) enqueueFrame(cmd byte, payload []byte) {
	_ = p.readQ.WriteByte(KISS_FEND)
	_ = p.readQ.WriteByte(cmd)
	if len(payload) > 0 {
		_, _ = p.readQ.Write(payload)
	}
	_ = p.readQ.WriteByte(KISS_FEND)
}

func (p *simSerialPort) handleFrame(cmd byte, data []byte) {
	data = kissUnescape(data)
	switch cmd {
	case KISS_CMD_DETECT:
		// request byte expected: DETECT_REQ
		p.enqueueFrame(KISS_CMD_DETECT, []byte{KISS_DETECT_RESP})
	case KISS_CMD_FW_VERSION:
		p.enqueueFrame(KISS_CMD_FW_VERSION, []byte{p.major, p.minor})
	case KISS_CMD_PLATFORM:
		p.enqueueFrame(KISS_CMD_PLATFORM, []byte{p.platform})
	case KISS_CMD_MCU:
		p.enqueueFrame(KISS_CMD_MCU, []byte{p.mcu})
	case KISS_CMD_BOARD:
		p.enqueueFrame(KISS_CMD_BOARD, []byte{p.board})
	case KISS_CMD_DEV_HASH:
		// request marker 0x01 in Detect() (ignored)
		_ = data
		p.enqueueFrame(KISS_CMD_DEV_HASH, p.deviceHash)
	case KISS_CMD_HASHES:
		if len(data) >= 1 && data[0] == 0x01 {
			p.enqueueFrame(KISS_CMD_HASHES, append([]byte{0x01}, p.firmwareHashTarget...))
		} else if len(data) >= 1 && data[0] == 0x02 {
			p.enqueueFrame(KISS_CMD_HASHES, append([]byte{0x02}, p.firmwareHash...))
		}
	case KISS_CMD_ROM_READ:
		// Return full EEPROM data (no escaping needed for our synthetic bytes).
		p.enqueueFrame(KISS_CMD_ROM_READ, p.eeprom)
	case KISS_CMD_CFG_READ:
		p.enqueueFrame(KISS_CMD_CFG_READ, p.cfgSector)
	case KISS_CMD_ROM_WRITE:
		// payload: [addr, value]
		if len(data) >= 2 {
			addr := int(data[0])
			val := data[1]
			if addr >= 0 && addr < len(p.eeprom) {
				p.eeprom[addr] = val
			}
		}
	case KISS_CMD_ROM_WIPE:
		// mark EEPROM as wiped/unprovisioned
		for i := range p.eeprom {
			p.eeprom[i] = 0x00
		}
		// keep size stable and clear lock byte
		if len(p.eeprom) > romAddrInfoLock {
			p.eeprom[romAddrInfoLock] = 0x00
		}
	case KISS_CMD_CONF_SAVE, KISS_CMD_CONF_DELETE, KISS_CMD_RESET, KISS_CMD_FW_UPD:
		// no-op acks
	case KISS_CMD_FW_HASH:
		// Set firmware hash.
		if len(data) == 32 {
			p.firmwareHash = append([]byte(nil), data...)
		}
	case KISS_CMD_DEV_SIG:
		// Store signature into EEPROM signature region if present.
		if len(data) == 128 && len(p.eeprom) >= int(romAddrSignature)+128 {
			copy(p.eeprom[romAddrSignature:romAddrSignature+128], data)
		}
	case KISS_CMD_WIFI_MODE, KISS_CMD_WIFI_CHN, KISS_CMD_WIFI_SSID, KISS_CMD_WIFI_PSK, KISS_CMD_WIFI_IP, KISS_CMD_WIFI_NM:
		// accept silently
	case KISS_CMD_BT_CTRL, KISS_CMD_BT_PIN, KISS_CMD_DISP_INT, KISS_CMD_DISP_BLNK, KISS_CMD_DISP_ROT, KISS_CMD_DISP_ADR, KISS_CMD_DISP_RCND, KISS_CMD_NP_INT, KISS_CMD_DIS_IA:
		// accept silently
	case KISS_CMD_LEAVE:
		// no-op
	default:
		// For development: surface unexpected commands.
		_ = fmt.Sprintf("simSerialPort: unhandled cmd 0x%02x", cmd)
	}
}

func kissUnescape(in []byte) []byte {
	if len(in) == 0 {
		return in
	}
	out := make([]byte, 0, len(in))
	esc := false
	for _, b := range in {
		if !esc {
			if b == KISS_FESC {
				esc = true
				continue
			}
			out = append(out, b)
			continue
		}
		switch b {
		case KISS_TFEND:
			out = append(out, KISS_FEND)
		case KISS_TFESC:
			out = append(out, KISS_FESC)
		default:
			out = append(out, b)
		}
		esc = false
	}
	return out
}
