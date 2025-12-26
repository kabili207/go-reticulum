package rns

import (
	"errors"
	"net"
)

// broadcastForInterface returns an IPv4 broadcast address for the named interface.
// Used for UDPInterface convenience config, similar to Python behavior.
func broadcastForInterface(name string) (net.IP, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	for _, a := range addrs {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP == nil || ipnet.Mask == nil {
			continue
		}
		ip4 := ipnet.IP.To4()
		if ip4 == nil {
			continue
		}
		bcast := make(net.IP, 4)
		for i := 0; i < 4; i++ {
			bcast[i] = ip4[i] | ^ipnet.Mask[i]
		}
		return bcast, nil
	}
	return nil, errors.New("no IPv4 address found on interface")
}
