//go:build !darwin && !linux && !windows

package ble

import "errors"

func New(string) (Transport, error) {
	return nil, errors.New("ble:// transport is unsupported on this platform in this Go port")
}

