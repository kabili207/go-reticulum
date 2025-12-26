package interfaces

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// ---- Framing ----
type HDLC struct{}

const (
	HDLC_FLAG     = 0x7E
	HDLC_ESC      = 0x7D
	HDLC_ESC_MASK = 0x20
)

func (HDLC) Escape(data []byte) []byte {
	out := make([]byte, 0, len(data)+8)
	for _, b := range data {
		switch b {
		case HDLC_ESC:
			out = append(out, HDLC_ESC, HDLC_ESC^HDLC_ESC_MASK)
		case HDLC_FLAG:
			out = append(out, HDLC_ESC, HDLC_FLAG^HDLC_ESC_MASK)
		default:
			out = append(out, b)
		}
	}
	return out
}

type KISS struct{}

const (
	KISS_FEND        = 0xC0
	KISS_FESC        = 0xDB
	KISS_TFEND       = 0xDC
	KISS_TFESC       = 0xDD
	KISS_CMD_DATA    = 0x00
	KISS_CMD_UNKNOWN = 0xFE
)

func (KISS) Escape(data []byte) []byte {
	out := make([]byte, 0, len(data)+8)
	for _, b := range data {
		switch b {
		case KISS_FESC:
			out = append(out, KISS_FESC, KISS_TFESC)
		case KISS_FEND:
			out = append(out, KISS_FESC, KISS_TFEND)
		default:
			out = append(out, b)
		}
	}
	return out
}

// ---- Common ----
type TCPLog interface {
	Debugf(string, ...any)
	Infof(string, ...any)
	Warnf(string, ...any)
	Errorf(string, ...any)
}

type TCPOwner interface {
	Inbound(data []byte, iface *TCPClientInterface)
}

const TCP_HW_MTU = 262144
const TCP_BITRATE_GUESS = 10 * 1000 * 1000
const TCP_DEFAULT_IFAC_SIZE = 16

// TCPTunnelSynthesizer can be set by the rns package (or callers) to mirror
// Python Transport.synthesize_tunnel(interface) behaviour for non-KISS TCP links.
// It is invoked after reconnecting a TCP initiator when WantsTunnel() is true.
var TCPTunnelSynthesizer func(iface *TCPClientInterface)

// ---- TCP Client Interface ----
type TCPClientInterface struct {
	Owner TCPOwner
	Log   TCPLog

	Name string

	TargetHost string
	TargetPort int

	KISSFraming bool
	I2PTunneled bool

	Initiator       bool
	ReconnectWait   time.Duration
	MaxReconnectTry *int // nil = infinite
	ConnectTimeout  time.Duration

	HWMTU    int
	Bitrate  int
	IFACSize int

	parent *TCPServerInterface

	mu         sync.Mutex
	conn       net.Conn
	writing    atomic.Bool
	online     atomic.Bool
	detached   atomic.Bool
	reconnect  atomic.Bool
	neverConn  atomic.Bool
	readOnce   sync.Once
	writeMutex sync.Mutex

	wantsTunnel atomic.Bool

	rxb atomic.Uint64
	txb atomic.Uint64
}

func NewTCPClientInitiator(owner TCPOwner, log TCPLog, name, host string, port int, kiss, i2p bool) *TCPClientInterface {
	iface := &TCPClientInterface{
		Owner:          owner,
		Log:            log,
		Name:           name,
		TargetHost:     host,
		TargetPort:     port,
		KISSFraming:    kiss,
		I2PTunneled:    i2p,
		Initiator:      true,
		ReconnectWait:  5 * time.Second,
		ConnectTimeout: 5 * time.Second,
		HWMTU:          TCP_HW_MTU,
		Bitrate:        TCP_BITRATE_GUESS,
		IFACSize:       TCP_DEFAULT_IFAC_SIZE,
	}
	iface.neverConn.Store(true)
	return iface
}

func NewTCPClientFromAccepted(owner TCPOwner, log TCPLog, name string, c net.Conn, kiss, i2p bool) *TCPClientInterface {
	iface := &TCPClientInterface{
		Owner:       owner,
		Log:         log,
		Name:        name,
		KISSFraming: kiss,
		I2PTunneled: i2p,
		Initiator:   false,
		HWMTU:       TCP_HW_MTU,
		Bitrate:     TCP_BITRATE_GUESS,
		IFACSize:    TCP_DEFAULT_IFAC_SIZE,
	}
	iface.setConn(c)
	iface.online.Store(true)
	iface.neverConn.Store(false)
	return iface
}

func (t *TCPClientInterface) String() string {
	host := t.TargetHost
	if host == "" && t.conn != nil {
		host = t.conn.RemoteAddr().String()
	}
	return fmt.Sprintf("TCPInterface[%s/%s:%d]", t.Name, host, t.TargetPort)
}

func (t *TCPClientInterface) setConn(c net.Conn) {
	t.mu.Lock()
	t.conn = c
	t.mu.Unlock()

	// Nagle off
	if tc, ok := c.(*net.TCPConn); ok {
		_ = tc.SetNoDelay(true)
		_ = tc.SetKeepAlive(true)
		// python: probe after ~5s (tcp keepidle / keepalive)
		_ = tc.SetKeepAlivePeriod(5 * time.Second)
	}

	// Best-effort: Linux TCP_USER_TIMEOUT + keepalive params
	_ = setTCPTimeoutsBestEffort(c, t.I2PTunneled)
}

func (t *TCPClientInterface) getConn() net.Conn {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.conn
}

func (t *TCPClientInterface) Detach() {
	t.online.Store(false)
	t.detached.Store(true)

	c := t.getConn()
	if c != nil {
		_ = c.Close()
	}
	t.mu.Lock()
	t.conn = nil
	t.mu.Unlock()
}

func (t *TCPClientInterface) InitialConnect(ctx context.Context) {
	ok := t.connect(ctx, true)
	if !ok {
		go t.reconnectLoop()
		return
	}
	go t.readLoop()
	// python: если не KISS, wants_tunnel = True (тут оставлено на уровень выше)
}

func (t *TCPClientInterface) connect(ctx context.Context, initial bool) bool {
	if t.TargetHost == "" || t.TargetPort == 0 {
		return false
	}
	addr := net.JoinHostPort(t.TargetHost, fmt.Sprintf("%d", t.TargetPort))
	if initial && t.Log != nil {
		t.Log.Debugf("Establishing TCP connection for %s...", t.String())
	}

	d := net.Dialer{Timeout: t.ConnectTimeout}
	c, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		if initial && t.Log != nil {
			t.Log.Errorf("Initial connection for %s could not be established: %v", t.String(), err)
			t.Log.Errorf("Leaving unconnected and retrying connection in %s.", t.ReconnectWait)
		}
		t.online.Store(false)
		return false
	}

	t.setConn(c)
	t.online.Store(true)
	t.writing.Store(false)
	t.neverConn.Store(false)

	if initial && t.Log != nil {
		t.Log.Debugf("TCP connection for %s established", t.String())
	}
	if !t.KISSFraming {
		t.wantsTunnel.Store(true)
	}
	return true
}

func (t *TCPClientInterface) reconnectLoop() {
	if !t.Initiator {
		if t.Log != nil {
			t.Log.Errorf("Attempt to reconnect on a non-initiator TCP interface: %s", t.String())
		}
		return
	}
	if !t.reconnect.CompareAndSwap(false, true) {
		return
	}
	defer t.reconnect.Store(false)

	attempts := 0
	for !t.online.Load() && !t.detached.Load() {
		time.Sleep(t.ReconnectWait)
		attempts++

		if t.MaxReconnectTry != nil && attempts > *t.MaxReconnectTry {
			if t.Log != nil {
				t.Log.Errorf("Max reconnection attempts reached for %s", t.String())
			}
			t.teardown()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), t.ConnectTimeout)
		ok := t.connect(ctx, false)
		cancel()
		if !ok {
			if t.Log != nil {
				t.Log.Debugf("Connection attempt for %s failed", t.String())
			}
			continue
		}
	}

	if !t.neverConn.Load() && t.Log != nil {
		t.Log.Infof("Reconnected socket for %s.", t.String())
	}

	// Python parity: if non-KISS framing is used, a tunnel may need to be
	// synthesized again after reconnect.
	if t.WantsTunnel() && TCPTunnelSynthesizer != nil {
		TCPTunnelSynthesizer(t)
	}

	go t.readLoop()
}

func (t *TCPClientInterface) ProcessIncoming(data []byte) {
	if !t.online.Load() || t.detached.Load() {
		return
	}
	t.rxb.Add(uint64(len(data)))
	if t.parent != nil {
		t.parent.rxb.Add(uint64(len(data)))
	}
	if t.Owner != nil && len(data) > 0 {
		cp := append([]byte(nil), data...)
		t.Owner.Inbound(cp, t)
	}
}

func (t *TCPClientInterface) ProcessOutgoing(data []byte) error {
	if !t.online.Load() || t.detached.Load() {
		return errors.New("tcp iface offline/detached")
	}

	c := t.getConn()
	if c == nil {
		return errors.New("tcp conn is nil")
	}

	var framed []byte
	if t.KISSFraming {
		framed = make([]byte, 0, len(data)+4)
		framed = append(framed, KISS_FEND, KISS_CMD_DATA)
		framed = append(framed, KISS{}.Escape(data)...)
		framed = append(framed, KISS_FEND)
	} else {
		framed = make([]byte, 0, len(data)+2+8)
		framed = append(framed, HDLC_FLAG)
		framed = append(framed, HDLC{}.Escape(data)...)
		framed = append(framed, HDLC_FLAG)
	}

	t.writeMutex.Lock()
	defer t.writeMutex.Unlock()

	t.writing.Store(true)
	_, err := c.Write(framed) // net.Conn.Write обычно пишет всё или ошибку
	t.writing.Store(false)

	if err != nil {
		if t.Log != nil {
			t.Log.Errorf("Exception occurred while transmitting via %s, tearing down interface", t.String())
			t.Log.Errorf("The contained exception was: %v", err)
		}
		t.teardown()
		return err
	}

	t.txb.Add(uint64(len(framed)))
	if t.parent != nil {
		t.parent.txb.Add(uint64(len(framed)))
	}
	return nil
}

func (t *TCPClientInterface) WantsTunnel() bool {
	return t.wantsTunnel.Load()
}

func (t *TCPClientInterface) readLoop() {
	// чтобы не запустить дважды
	called := false
	t.readOnce.Do(func() { called = true })
	if !called {
		return
	}

	defer func() {
		t.online.Store(false)
	}()

	c := t.getConn()
	if c == nil {
		return
	}

	buf := make([]byte, 4096)

	// --- KISS state ---
	inFrame := false
	escape := false
	command := byte(KISS_CMD_UNKNOWN)
	var dataBuf bytes.Buffer

	// --- HDLC buffer ---
	var frameBuf bytes.Buffer

	for {
		n, err := c.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) || t.detached.Load() {
				// как python: socket closed
			} else if t.Log != nil {
				t.Log.Warnf("An interface error occurred for %s: %v", t.String(), err)
			}

			t.online.Store(false)
			if t.Initiator && !t.detached.Load() {
				if t.Log != nil {
					t.Log.Warnf("The socket for %s was closed, attempting to reconnect...", t.String())
				}
				t.reconnectLoop()
			} else {
				if t.Log != nil {
					t.Log.Debugf("The socket for remote client %s was closed.", t.String())
				}
				t.teardown()
			}
			return
		}

		if n == 0 {
			t.online.Store(false)
			if t.Initiator && !t.detached.Load() {
				if t.Log != nil {
					t.Log.Warnf("The socket for %s was closed, attempting to reconnect...", t.String())
				}
				t.reconnectLoop()
			} else {
				t.teardown()
			}
			return
		}

		chunk := buf[:n]

		if t.KISSFraming {
			// bytewise parser, как в python
			for i := 0; i < len(chunk); i++ {
				b := chunk[i]
				if inFrame && b == KISS_FEND && command == KISS_CMD_DATA {
					inFrame = false
					t.ProcessIncoming(dataBuf.Bytes())
					dataBuf.Reset()
					continue
				} else if b == KISS_FEND {
					inFrame = true
					command = KISS_CMD_UNKNOWN
					dataBuf.Reset()
					escape = false
					continue
				} else if inFrame && dataBuf.Len() < t.HWMTU {
					if dataBuf.Len() == 0 && command == KISS_CMD_UNKNOWN {
						// strip port nibble
						command = b & 0x0F
						continue
					}
					if command == KISS_CMD_DATA {
						if b == KISS_FESC {
							escape = true
							continue
						}
						if escape {
							if b == KISS_TFEND {
								b = KISS_FEND
							} else if b == KISS_TFESC {
								b = KISS_FESC
							}
							escape = false
						}
						dataBuf.WriteByte(b)
					}
				}
			}
		} else {
			// HDLC framing: накапливаем и вырезаем FLAG..FLAG
			frameBuf.Write(chunk)

			for {
				all := frameBuf.Bytes()
				start := bytes.IndexByte(all, HDLC_FLAG)
				if start < 0 {
					// ничего полезного
					frameBuf.Reset()
					break
				}
				end := bytes.IndexByte(all[start+1:], HDLC_FLAG)
				if end < 0 {
					// ждём второй флаг
					if start > 0 {
						frameBuf.Next(start) // выброс мусора до первого флага
					}
					break
				}
				end = start + 1 + end
				frame := append([]byte(nil), all[start+1:end]...)

				// unescape как в python replace-парами
				frame = bytes.ReplaceAll(frame, []byte{HDLC_ESC, HDLC_FLAG ^ HDLC_ESC_MASK}, []byte{HDLC_FLAG})
				frame = bytes.ReplaceAll(frame, []byte{HDLC_ESC, HDLC_ESC ^ HDLC_ESC_MASK}, []byte{HDLC_ESC})

				// python: если len(frame) > HEADER_MINSIZE -> process
				// тут оставляю >0, а проверку заголовка делай выше по стеку
				if len(frame) > 0 {
					t.ProcessIncoming(frame)
				}

				// отрезаем до end (оставляем end как новый старт, как python frame_buffer = frame_buffer[frame_end:])
				frameBuf.Next(end)
			}
		}
	}
}

func (t *TCPClientInterface) teardown() {
	t.online.Store(false)
	c := t.getConn()
	if c != nil {
		_ = c.Close()
	}
	t.mu.Lock()
	t.conn = nil
	t.mu.Unlock()

	// если spawned у сервера — убрать из списка
	if t.parent != nil {
		t.parent.removeClient(t)
	}
}

// ---- TCP Server Interface ----
type TCPServerInterface struct {
	Owner TCPOwner
	Log   TCPLog

	Name string

	ListenAddr  string // "ip:port" или "[ip%if]:port"
	I2PTunneled bool
	KISSFraming bool

	HWMTU int

	ln       net.Listener
	online   atomic.Bool
	detached atomic.Bool

	mu      sync.Mutex
	clients []*TCPClientInterface

	rxb atomic.Uint64
	txb atomic.Uint64
}

func NewTCPServer(owner TCPOwner, log TCPLog, name, listenAddr string, kiss, i2p bool) *TCPServerInterface {
	return &TCPServerInterface{
		Owner:       owner,
		Log:         log,
		Name:        name,
		ListenAddr:  listenAddr,
		I2PTunneled: i2p,
		KISSFraming: kiss,
		HWMTU:       TCP_HW_MTU,
	}
}

func (s *TCPServerInterface) String() string {
	return fmt.Sprintf("TCPServerInterface[%s/%s]", s.Name, s.ListenAddr)
}

func (s *TCPServerInterface) Clients() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.clients)
}

func (s *TCPServerInterface) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	s.ln = ln
	s.online.Store(true)

	if s.Log != nil {
		s.Log.Infof("Listening on %s", s.String())
	}

	go s.acceptLoop()
	return nil
}

func (s *TCPServerInterface) Detach() {
	s.detached.Store(true)
	s.online.Store(false)
	if s.ln != nil {
		_ = s.ln.Close()
	}
	// закрыть клиентов
	s.mu.Lock()
	cs := append([]*TCPClientInterface(nil), s.clients...)
	s.clients = nil
	s.mu.Unlock()

	for _, c := range cs {
		c.Detach()
	}
}

func (s *TCPServerInterface) acceptLoop() {
	for s.online.Load() && !s.detached.Load() {
		c, err := s.ln.Accept()
		if err != nil {
			if s.detached.Load() {
				return
			}
			if s.Log != nil {
				s.Log.Warnf("Accept error on %s: %v", s.String(), err)
			}
			continue
		}

		if s.Log != nil {
			s.Log.Debugf("Accepting incoming TCP connection")
		}

		ra := c.RemoteAddr().(*net.TCPAddr)
		iface := NewTCPClientFromAccepted(s.Owner, s.Log, "Client on "+s.Name, c, s.KISSFraming, s.I2PTunneled)
		iface.parent = s
		iface.TargetHost = ra.IP.String()
		iface.TargetPort = ra.Port
		iface.HWMTU = s.HWMTU

		s.mu.Lock()
		s.clients = append(s.clients, iface)
		s.mu.Unlock()

		go iface.readLoop()
	}
}

func (s *TCPServerInterface) removeClient(ci *TCPClientInterface) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.clients[:0]
	for _, c := range s.clients {
		if c != ci {
			out = append(out, c)
		}
	}
	s.clients = out
}

// ---- Best-effort socket options ----
