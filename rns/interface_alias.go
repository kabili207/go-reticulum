package rns

import ifaces "github.com/svanichkin/go-reticulum/rns/interfaces"
import "time"

// Interface is re-exported from the interfaces subpackage so Go ports can live next to
// the original Python drivers in rns/interfaces/.
type Interface = ifaces.Interface

func init() {
	// Provide inbound callback to interfaces package without introducing import cycles.
	ifaces.InboundHandler = func(raw []byte, ifc *ifaces.Interface) {
		Inbound(raw, ifc)
	}
	ifaces.DiagLog = Log
	ifaces.DiagLogf = Logf
	ifaces.ExitFunc = Exit
	ifaces.PanicFunc = Panic
	ifaces.PanicOnInterfaceErrorProvider = func() bool {
		r := GetInstance()
		if r == nil {
			return false
		}
		return r.PanicOnInterfaceError
	}
	ifaces.SpawnHandler = func(ifc *ifaces.Interface) {
		// Register spawned sub-interfaces like AutoInterface peers.
		AddInterface(ifc)
	}
	ifaces.RemoveInterfaceHandler = removeInterface
	ifaces.QueuedAnnounceLife = time.Duration(QUEUED_ANNOUNCE_LIFE) * time.Second
	ifaces.HeaderMinSize = HEADER_MINSIZE
	ifaces.TransportIdentityHashProvider = func() []byte {
		if TransportIdentity == nil {
			return nil
		}
		return TransportIdentity.Hash
	}
	ifaces.TunnelSynthesizer = func(ifc *ifaces.Interface) {
		SynthesizeTunnel(ifc)
	}
}
