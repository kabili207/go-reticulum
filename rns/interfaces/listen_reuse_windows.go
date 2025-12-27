//go:build windows

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
			// Match Python LocalInterface behaviour on Windows: SO_EXCLUSIVEADDRUSE.
			_ = syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_EXCLUSIVEADDRUSE, 1)
		})
		return firstErr
	}
	return lc.Listen(context.Background(), network, address)
}

