//go:build darwin || linux

package main

import (
	"errors"
	"os"
	"strings"
	"syscall"
)

type fileSerialPort struct {
	f    *os.File
	name string
}

func OpenSerialPort(path string) (SerialPort, error) {
	if strings.HasPrefix(path, "sim://") {
		return newSimSerialPort(path)
	}
	f, err := os.OpenFile(path, os.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	p := &fileSerialPort{f: f, name: path}
	if err := configureSerialPort(f.Fd()); err != nil {
		_ = f.Close()
		return nil, err
	}
	_ = p.SetDTR(true)
	_ = p.SetRTS(true)
	return p, nil
}

func (p *fileSerialPort) Read(b []byte) (int, error) {
	n, err := p.f.Read(b)
	if err == nil {
		return n, nil
	}
	if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
		return 0, nil
	}
	return n, err
}
func (p *fileSerialPort) Write(b []byte) (int, error) { return p.f.Write(b) }
func (p *fileSerialPort) Close() error                { return p.f.Close() }
func (p *fileSerialPort) IsOpen() bool                { return p != nil && p.f != nil }
func (p *fileSerialPort) Name() string                { return p.name }

func (p *fileSerialPort) BytesAvailable() (int, error) {
	// Best-effort. Some platforms don't expose FIONREAD in syscall.
	// The read loop in this utility can operate without this hint.
	return 0, nil
}
