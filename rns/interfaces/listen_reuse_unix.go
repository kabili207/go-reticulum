//go:build linux || darwin

package interfaces

import (
	"context"
	"net"
	"syscall"
)

func listenWithReuseAddr(network, address string) (net.Listener, error) {
	var lc net.ListenConfig
	lc.Control = func(_, _ string, c syscall.RawConn) error {
		var firstErr error
		_ = c.Control(func(fd uintptr) {
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil && firstErr == nil {
				firstErr = err
			}
		})
		return firstErr
	}
	return lc.Listen(context.Background(), network, address)
}

