package interfaces

import (
	"context"

	blepkg "github.com/svanichkin/go-reticulum/rns/interfaces/ble"
)

type bleTransportAdapter struct {
	inner blepkg.Transport
}

func (b *bleTransportAdapter) Open(ctx context.Context) error { return b.inner.Open(ctx) }
func (b *bleTransportAdapter) Close() error                   { return b.inner.Close() }
func (b *bleTransportAdapter) Read(p []byte) (int, error)     { return b.inner.Read(p) }
func (b *bleTransportAdapter) Write(p []byte) (int, error)    { return b.inner.Write(p) }
func (b *bleTransportAdapter) IsOpen() bool                   { return b.inner.IsOpen() }
func (b *bleTransportAdapter) DeviceDisappeared() bool        { return b.inner.DeviceDisappeared() }
func (b *bleTransportAdapter) String() string                 { return b.inner.String() }

func newBLETransport(target string) (Transport, error) {
	t, err := blepkg.New(target)
	if err != nil {
		return nil, err
	}
	return &bleTransportAdapter{inner: t}, nil
}
