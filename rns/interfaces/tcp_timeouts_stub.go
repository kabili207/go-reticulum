//go:build !linux && !darwin

package interfaces

import "net"

func setTCPTimeoutsBestEffort(_ net.Conn, _ bool) error {
	return nil
}
