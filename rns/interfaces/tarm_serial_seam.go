package interfaces

import (
	"time"

	"github.com/tarm/serial"
)

// tarmSerialPort is the minimal interface we need from github.com/tarm/serial.Port.
// It allows unit/integration tests to substitute a fake serial port.
type tarmSerialPort interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close() error
}

var openTarmSerialPort = func(cfg *serial.Config) (tarmSerialPort, error) {
	return serial.OpenPort(cfg)
}

var ax25Sleep = func(dur time.Duration) { time.Sleep(dur) }

var kissSleep = func(dur time.Duration) { time.Sleep(dur) }
