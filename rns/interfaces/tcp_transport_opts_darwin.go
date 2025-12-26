//go:build darwin

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

	const tcpProbeAfterS = 5

	var firstErr error
	_ = raw.Control(func(fd uintptr) {
		if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1); err != nil && firstErr == nil {
			firstErr = err
		}
		// Darwin uses TCP_KEEPALIVE for idle seconds (like Python).
		if err := unix.SetsockoptInt(int(fd), unix.IPPROTO_TCP, unix.TCP_KEEPALIVE, tcpProbeAfterS); err != nil && firstErr == nil {
			firstErr = err
		}
	})
	return firstErr
}
