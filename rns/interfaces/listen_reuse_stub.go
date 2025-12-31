//go:build !linux && !darwin && !windows

package interfaces

import "net"

func listenWithReuseAddr(network, address string) (net.Listener, error) {
	return net.Listen(network, address)
}
