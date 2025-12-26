//go:build !darwin && !linux

package main

import "errors"

func OpenSerialPort(_ string) (SerialPort, error) {
	return nil, errors.New("serial port support is not implemented on this platform")
}

