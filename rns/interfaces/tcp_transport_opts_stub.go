//go:build !linux && !darwin

package interfaces

import (
	"net"
)

// setTCPPlatformOptions applies best-effort socket options to match Python TCPConnection.
// Stubbed on unsupported platforms.
func setTCPPlatformOptions(_ *net.TCPConn) error { return nil }
