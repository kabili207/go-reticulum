package interfaces

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	pipeHWMTU        = 1064
	pipeBitrateGuess = 1_000_000
)

type pipeConfig struct {
	Name         string
	Command      string
	RespawnDelay time.Duration
}

type PipeDriver struct {
	iface *Interface
	cfg   pipeConfig

	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser

	online atomic.Bool
	stopCh chan struct{}

	writeMu sync.Mutex
}

func NewPipeInterface(name string, kv map[string]string) (*Interface, error) {
	cfg, err := parsePipeConfig(name, kv)
	if err != nil {
		return nil, err
	}

	iface := &Interface{
		Name:              name,
		Type:              "PipeInterface",
		IN:                true,
		OUT:               true,
		DriverImplemented: true,
		Bitrate:           pipeBitrateGuess,
		HWMTU:             pipeHWMTU,
		Online:            false,
		Created:           time.Now(),
	}
	// Match Python PipeInterface.DEFAULT_IFAC_SIZE.
	if iface.IFACSize == 0 {
		iface.IFACSize = 8
	}
	driver := &PipeDriver{
		iface:  iface,
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
	iface.pipe = driver

	if err := driver.openAndConfigure(); err != nil {
		return iface, err
	}
	return iface, nil
}

func parsePipeConfig(name string, kv map[string]string) (pipeConfig, error) {
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

	cfg := pipeConfig{
		Name:         name,
		Command:      get("command"),
		RespawnDelay: 5 * time.Second,
	}
	if strings.TrimSpace(cfg.Command) == "" {
		return cfg, fmt.Errorf("no command specified for PipeInterface %q", name)
	}
	if s := get("respawn_delay"); s != "" {
		if f, err := strconv.ParseFloat(s, 64); err == nil && f > 0 {
			cfg.RespawnDelay = time.Duration(f * float64(time.Second))
		}
	}
	if cfg.RespawnDelay <= 0 {
		cfg.RespawnDelay = 5 * time.Second
	}
	return cfg, nil
}

func (p *PipeDriver) String() string {
	return "PipeInterface[" + p.cfg.Name + "]"
}

func (p *PipeDriver) Close() {
	select {
	case <-p.stopCh:
		return
	default:
		close(p.stopCh)
	}
	p.iface.Online = false
	p.online.Store(false)
	_ = p.killProcess()
}

func (p *PipeDriver) ProcessOutgoing(data []byte) {
	if p == nil || len(data) == 0 || p.iface == nil || !p.online.Load() {
		return
	}

	// Python does not truncate outgoing payload before framing; receiver-side HWMTU caps.
	framed := make([]byte, 0, len(data)+2+8)
	framed = append(framed, hdlcFlag)
	framed = append(framed, hdlcEscape(data)...)
	framed = append(framed, hdlcFlag)

	p.writeMu.Lock()
	defer p.writeMu.Unlock()

	p.mu.Lock()
	stdin := p.stdin
	p.mu.Unlock()
	if stdin == nil {
		return
	}

	n, err := stdin.Write(framed)
	if err != nil || n != len(framed) {
		if DiagLogf != nil {
			DiagLogf(LogError, "Exception occurred while transmitting via %s, tearing down interface", p.iface)
			if err != nil {
				DiagLogf(LogError, "The contained exception was: %v", err)
			} else {
				DiagLogf(LogError, "The contained exception was: short write (%d/%d)", n, len(framed))
			}
		}
		p.iface.Online = false
		p.online.Store(false)
		_ = p.killProcess()
		go p.reconnectLoop()
		return
	}

	atomic.AddUint64(&p.iface.TXB, uint64(len(framed)))
	if parent := p.iface.Parent; parent != nil {
		atomic.AddUint64(&parent.TXB, uint64(len(framed)))
	}
}

func (p *PipeDriver) openAndConfigure() error {
	if DiagLogf != nil {
		DiagLogf(LogVerbose, "Connecting subprocess pipe for %s...", p.iface)
	}
	if err := p.openPipe(); err != nil {
		if DiagLogf != nil {
			DiagLogf(LogError, "Could connect pipe for interface %s", p.iface)
		}
		return err
	}
	time.Sleep(10 * time.Millisecond)
	p.iface.Online = true
	p.online.Store(true)
	go p.readLoop()
	if DiagLogf != nil {
		DiagLogf(LogVerbose, "Subprocess pipe for %s is now connected", p.iface)
	}
	return nil
}

func shlexSplit(s string) ([]string, error) {
	// Minimal shlex.split compatible with Python defaults:
	// - supports single and double quotes
	// - supports backslash escaping outside single quotes
	var args []string
	var cur strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if cur.Len() > 0 {
			args = append(args, cur.String())
			cur.Reset()
		}
	}

	for _, r := range s {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\' && !inSingle:
			escaped = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case (r == ' ' || r == '\t' || r == '\n' || r == '\r') && !inSingle && !inDouble:
			flush()
		default:
			cur.WriteRune(r)
		}
	}
	if escaped || inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quotes/escape in command")
	}
	flush()
	if len(args) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	return args, nil
}

func (p *PipeDriver) openPipe() error {
	args, err := shlexSplit(p.cfg.Command)
	if err != nil {
		return err
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return err
	}
	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return err
	}

	p.mu.Lock()
	p.cmd = cmd
	p.stdin = stdin
	p.stdout = stdout
	p.mu.Unlock()
	return nil
}

func (p *PipeDriver) processIncoming(data []byte) {
	if p == nil || p.iface == nil || len(data) == 0 {
		return
	}
	atomic.AddUint64(&p.iface.RXB, uint64(len(data)))
	if parent := p.iface.Parent; parent != nil {
		atomic.AddUint64(&parent.RXB, uint64(len(data)))
	}
	if InboundHandler != nil {
		InboundHandler(append([]byte(nil), data...), p.iface)
	}
}

func (p *PipeDriver) readLoop() {
	defer func() {
		p.iface.Online = false
		p.online.Store(false)
		go p.reconnectLoop()
	}()

	p.mu.Lock()
	stdout := p.stdout
	p.mu.Unlock()
	if stdout == nil {
		return
	}

	r := bufio.NewReaderSize(stdout, 4096)

	inFrame := false
	escape := false
	buf := make([]byte, 0, pipeHWMTU)
	var lastReadMS int64
	var readErr error
	for {
		select {
		case <-p.stopCh:
			return
		default:
		}

		b, err := r.ReadByte()
		if err != nil {
			readErr = err
			break
		}
		lastReadMS = time.Now().UnixMilli()

		if inFrame && b == hdlcFlag {
			inFrame = false
			p.processIncoming(buf)
			buf = buf[:0]
			continue
		}
		if b == hdlcFlag {
			inFrame = true
			escape = false
			buf = buf[:0]
			continue
		}
		if inFrame && len(buf) < pipeHWMTU {
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
			buf = append(buf, b)
		}
	}
	_ = lastReadMS

	if readErr != nil && !errors.Is(readErr, io.EOF) {
		if DiagLogf != nil {
			DiagLogf(LogError, "A pipe error occurred, the contained exception was: %v", readErr)
			DiagLogf(LogError, "The interface %s experienced an unrecoverable error and is now offline.", p.iface)
		}
		_ = p.killProcess()
		if PanicOnInterfaceErrorProvider != nil && PanicOnInterfaceErrorProvider() {
			if PanicFunc != nil {
				PanicFunc()
			}
		}
		if DiagLogf != nil {
			DiagLogf(LogError, "Reticulum will attempt to reconnect the interface periodically.")
		}
		return
	}

	if DiagLogf != nil {
		DiagLogf(LogError, "Subprocess terminated on %s", p.iface)
	}
	_ = p.killProcess()
}

func (p *PipeDriver) killProcess() error {
	p.mu.Lock()
	cmd := p.cmd
	stdin := p.stdin
	stdout := p.stdout
	p.cmd = nil
	p.stdin = nil
	p.stdout = nil
	p.mu.Unlock()

	if stdin != nil {
		_ = stdin.Close()
	}
	if stdout != nil {
		_ = stdout.Close()
	}
	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}
	return nil
}

func (p *PipeDriver) reconnectLoop() {
	for {
		select {
		case <-p.stopCh:
			return
		case <-time.After(p.cfg.RespawnDelay):
		}

		if p.online.Load() {
			return
		}

		if DiagLogf != nil {
			DiagLogf(LogVerbose, "Attempting to respawn subprocess for %s...", p.iface)
		}

		if err := p.openAndConfigure(); err != nil {
			if DiagLogf != nil {
				DiagLogf(LogError, "Error while spawning subprocess, the contained exception was: %v", err)
			}
			continue
		}

		if DiagLogf != nil {
			DiagLogf(LogVerbose, "Reconnected pipe for %s", p.iface)
		}
		return
	}
}
