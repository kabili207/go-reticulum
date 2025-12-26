//go:build darwin

package main

import (
	// uses globSerialPatterns from serial_ports_common.go
)

func listSerialPorts() ([]string, error) {
	patterns := []string{
		"/dev/cu.*",
		"/dev/tty.*",
	}
	return globSerialPatterns(patterns), nil
}
