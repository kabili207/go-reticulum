//go:build linux

package main

import (
	"path/filepath"
)

func listSerialPorts() ([]string, error) {
	patterns := []string{
		"/dev/serial/by-id/*",
		"/dev/ttyUSB*",
		"/dev/ttyACM*",
		"/dev/ttyAMA*",
		"/dev/ttyS*",
	}
	return globSerialPatterns(patterns), nil
}

