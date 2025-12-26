//go:build linux

package interfaces

import (
	"net"

	"golang.org/x/sys/unix"
)

func setTCPPlatformOptions(conn *net.TCPConn) error {
	if conn == nil {
		return nil
	}
	raw, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	const (
		tcpUserTimeoutMS = 24 * 1000
		tcpProbeAfterS   = 5
		tcpProbeIntvlS   = 2
		tcpProbes        = 12
	)

	var firstErr error
	_ = raw.Control(func(fd uintptr) {
		if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_USER_TIMEOUT, tcpUserTimeoutMS); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPIDLE, tcpProbeAfterS); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPINTVL, tcpProbeIntvlS); err != nil && firstErr == nil {
			firstErr = err
		}
		if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPCNT, tcpProbes); err != nil && firstErr == nil {
			firstErr = err
		}
	})
	return firstErr
}
