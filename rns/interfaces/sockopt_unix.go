//go:build !windows

package interfaces

import "syscall"

func setSockoptIntFD(fd uintptr, level, opt, value int) error {
	return syscall.SetsockoptInt(int(fd), level, opt, value)
}

func setSockoptIPv6MreqFD(fd uintptr, level, opt int, mreq *syscall.IPv6Mreq) error {
	return syscall.SetsockoptIPv6Mreq(int(fd), level, opt, mreq)
}

