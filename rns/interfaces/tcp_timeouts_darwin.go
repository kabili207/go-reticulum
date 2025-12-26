//go:build darwin

package interfaces

import (
	"net"
	"syscall"
)

func setTCPTimeoutsBestEffort(c net.Conn, i2p bool) error {
	tc, ok := c.(*net.TCPConn)
	if !ok {
		return nil
	}
	raw, err := tc.SyscallConn()
	if err != nil {
		return nil
	}

	probeAfter := 5
	if i2p {
		probeAfter = 10
	}

	var serr error
	_ = raw.Control(func(fd uintptr) {
		_ = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1)
		const TCP_KEEPIDLE_DARWIN = 0x10
		_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, TCP_KEEPIDLE_DARWIN, probeAfter)
	})
	return serr
}
