//go:build ios

package interfaces

import (
	"context"

	blepkg "github.com/svanichkin/go-reticulum/rns/interfaces/ble"
)

func newBLETransport(target string) (Transport, error) {
	// On iOS we provide a CoreBluetooth-backed BLE UART transport in rns/interfaces/ble.
	t, err := blepkg.New(target)
	if err != nil {
		return nil, err
	}
	return &bleTransportAdapterIOS{inner: t}, nil
}

type bleTransportAdapterIOS struct {
	inner blepkg.Transport
}

func (b *bleTransportAdapterIOS) Open(ctx context.Context) error { return b.inner.Open(ctx) }
func (b *bleTransportAdapterIOS) Close() error                   { return b.inner.Close() }
func (b *bleTransportAdapterIOS) Read(p []byte) (int, error)     { return b.inner.Read(p) }
func (b *bleTransportAdapterIOS) Write(p []byte) (int, error)    { return b.inner.Write(p) }
func (b *bleTransportAdapterIOS) IsOpen() bool                   { return b.inner.IsOpen() }
func (b *bleTransportAdapterIOS) DeviceDisappeared() bool        { return b.inner.DeviceDisappeared() }
func (b *bleTransportAdapterIOS) String() string                 { return b.inner.String() }
