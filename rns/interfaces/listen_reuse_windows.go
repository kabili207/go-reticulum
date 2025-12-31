//go:build windows

package interfaces

import (
	"context"
	"net"
	"syscall"

	"golang.org/x/sys/windows"
)

func listenWithReuseAddr(network, address string) (net.Listener, error) {
	var lc net.ListenConfig
	lc.Control = func(_, _ string, c syscall.RawConn) error {
		var firstErr error
		_ = c.Control(func(fd uintptr) {
			// Match Python LocalInterface behaviour on Windows: SO_EXCLUSIVEADDRUSE.
			_ = windows.SetsockoptInt(windows.Handle(fd), windows.SOL_SOCKET, windows.SO_EXCLUSIVEADDRUSE, 1)
		})
		return firstErr
	}
	return lc.Listen(context.Background(), network, address)
}
