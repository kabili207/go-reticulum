//go:build linux

package main

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

func configureSerialPort(fd uintptr) error {
	var tio syscall.Termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&tio)))
	if errno != 0 {
		return fmt.Errorf("ioctl(TCGETS): %v", errno)
	}

	// Raw 8N1, like pyserial defaults used in rnodeconf.py.
	tio.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	tio.Oflag &^= syscall.OPOST
	tio.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	tio.Cflag &^= syscall.CSIZE | syscall.PARENB
	tio.Cflag |= syscall.CS8 | syscall.CLOCAL | syscall.CREAD

	// Non-blocking reads (timeout=0 in Python).
	tio.Cc[syscall.VMIN] = 0
	tio.Cc[syscall.VTIME] = 0

	// 115200 baud (rnodeconf.py rnode_baudrate).
	tio.Ispeed = syscall.B115200
	tio.Ospeed = syscall.B115200

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&tio)))
	if errno != 0 {
		return fmt.Errorf("ioctl(TCSETS): %v", errno)
	}
	return nil
}

func (p *fileSerialPort) SetDTR(on bool) error { return setModemControlLine(p, syscall.TIOCM_DTR, on) }
func (p *fileSerialPort) SetRTS(on bool) error { return setModemControlLine(p, syscall.TIOCM_RTS, on) }

func setModemControlLine(p *fileSerialPort, bit int, on bool) error {
	if p == nil || p.f == nil {
		return errors.New("serial port is not open")
	}
	fd := p.f.Fd()
	var status int
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCMGET), uintptr(unsafe.Pointer(&status)))
	if errno != 0 {
		return fmt.Errorf("ioctl(TIOCMGET): %v", errno)
	}
	if on {
		status |= bit
	} else {
		status &^= bit
	}
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCMSET), uintptr(unsafe.Pointer(&status)))
	if errno != 0 {
		return fmt.Errorf("ioctl(TIOCMSET): %v", errno)
	}
	return nil
}

