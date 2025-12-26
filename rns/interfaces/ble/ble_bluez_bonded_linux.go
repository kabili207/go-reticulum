//go:build linux

package ble

import (
	"errors"
	"fmt"

	"github.com/godbus/dbus/v5"
)

func bluezIsDeviceBonded(addr string) (bool, error) {
	if addr == "" {
		return false, errors.New("missing device address")
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		return false, err
	}
	defer conn.Close()

	const (
		bluezBusName = "org.bluez"
		objMgrIface  = "org.freedesktop.DBus.ObjectManager"
		deviceIface  = "org.bluez.Device1"
		propsIface   = "org.freedesktop.DBus.Properties"
	)

	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	obj := conn.Object(bluezBusName, "/")
	if err := obj.Call(objMgrIface+".GetManagedObjects", 0).Store(&objects); err != nil {
		return false, err
	}

	for path, ifaces := range objects {
		props, ok := ifaces[deviceIface]
		if !ok {
			continue
		}
		v, ok := props["Address"]
		if !ok {
			continue
		}
		devAddr, ok := v.Value().(string)
		if !ok || devAddr == "" {
			continue
		}
		if devAddr != addr {
			continue
		}

		var paired bool
		if err := conn.Object(bluezBusName, path).Call(propsIface+".Get", 0, deviceIface, "Paired").Store(&paired); err != nil {
			return false, err
		}
		return paired, nil
	}

	return false, fmt.Errorf("bluez device not found: %s", addr)
}

