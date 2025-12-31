//go:build linux

package interfaces

import (
	"net"

	"golang.org/x/sys/unix"
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

	userTimeout := 24
	probeAfter := 5
	probeInterval := 2
	probes := 12
	if i2p {
		userTimeout = 45
		probeAfter = 10
		probeInterval = 9
		probes = 5
	}

	var serr error
	_ = raw.Control(func(fd uintptr) {
		_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_USER_TIMEOUT, userTimeout*1000)
		_ = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1)
		_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, probeAfter)
		_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, probeInterval)
		_ = unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPCNT, probes)
	})
	return serr
}
