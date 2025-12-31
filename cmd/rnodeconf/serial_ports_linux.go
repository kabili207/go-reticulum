//go:build linux

package main

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
