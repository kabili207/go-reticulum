package interfaces

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func tcpAddressForHost(host string, port int, preferIPv6 bool) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", fmt.Errorf("empty host")
	}
	if port <= 0 || port > 65535 {
		return "", fmt.Errorf("invalid port: %d", port)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		// Keep original host, it may already be a literal or include a scope.
		return net.JoinHostPort(host, strconv.Itoa(port)), nil
	}
	if len(ips) == 0 {
		return "", fmt.Errorf("no addresses for host %q", host)
	}

	var chosen net.IP
	for _, ip := range ips {
		if preferIPv6 && ip.To4() == nil {
			chosen = ip
			break
		}
		if !preferIPv6 && ip.To4() != nil {
			chosen = ip
			break
		}
	}
	if chosen == nil {
		chosen = ips[0]
	}
	return net.JoinHostPort(chosen.String(), strconv.Itoa(port)), nil
}

func tcpAddressForInterface(ifName string, port int, preferIPv6 bool) (string, error) {
	ifName = strings.TrimSpace(ifName)
	if ifName == "" {
		return "", fmt.Errorf("empty interface name")
	}
	if port <= 0 || port > 65535 {
		return "", fmt.Errorf("invalid port: %d", port)
	}

	ifc, err := net.InterfaceByName(ifName)
	if err != nil {
		return "", err
	}
	addrs, err := ifc.Addrs()
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("no addresses for interface %q", ifName)
	}

	var chosenIP net.IP
	var zone string
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP
		if ip == nil || ip.IsUnspecified() || ip.IsLoopback() {
			continue
		}
		if preferIPv6 && ip.To4() == nil {
			chosenIP = ip
			break
		}
		if !preferIPv6 && ip.To4() != nil {
			chosenIP = ip
			break
		}
		if chosenIP == nil {
			chosenIP = ip
		}
	}
	if chosenIP == nil {
		return "", fmt.Errorf("no usable addresses for interface %q", ifName)
	}

	// Link-local IPv6 needs a zone (scope id) for correct binding.
	if chosenIP.To4() == nil && strings.HasPrefix(strings.ToLower(chosenIP.String()), "fe80:") {
		zone = ifName
	}

	if zone != "" {
		return net.JoinHostPort(chosenIP.String()+"%"+zone, strconv.Itoa(port)), nil
	}
	return net.JoinHostPort(chosenIP.String(), strconv.Itoa(port)), nil
}

