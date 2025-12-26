//go:build linux

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
		_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_USER_TIMEOUT, userTimeout*1000)
		_ = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1)
		_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, probeAfter)
		_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, probeInterval)
		_ = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, probes)
	})
	return serr
}
