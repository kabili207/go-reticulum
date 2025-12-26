//go:build !linux

package ble

func bluezIsDeviceBonded(_ string) (bool, error) { return true, nil }

