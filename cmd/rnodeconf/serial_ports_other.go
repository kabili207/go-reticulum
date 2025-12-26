//go:build !darwin && !linux

package main

func listSerialPorts() ([]string, error) {
	return nil, nil
}

