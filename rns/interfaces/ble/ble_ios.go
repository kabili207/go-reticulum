//go:build ios

package ble

/*
#cgo CFLAGS: -fobjc-arc
#cgo LDFLAGS: -framework Foundation -framework CoreBluetooth

#include <stdint.h>
#include <stdlib.h>

typedef void* reticulum_ble_t;

reticulum_ble_t reticulum_ble_new(const char* target_name, const char* target_addr);
int reticulum_ble_open(reticulum_ble_t h, double timeout_seconds);
void reticulum_ble_close(reticulum_ble_t h);
void reticulum_ble_free(reticulum_ble_t h);
int reticulum_ble_is_open(reticulum_ble_t h);
int reticulum_ble_device_disappeared(reticulum_ble_t h);
int reticulum_ble_read(reticulum_ble_t h, uint8_t* out, int out_cap);
int reticulum_ble_write(reticulum_ble_t h, const uint8_t* data, int data_len);
const char* reticulum_ble_debug_string(reticulum_ble_t h);
*/
import "C"

import (
	"context"
	"errors"
	"io"
	"unsafe"
)

// iosTransport is a CoreBluetooth-backed implementation of the Reticulum BLE UART transport.
// It matches the same NUS UUIDs used by other platforms:
// - Service: 6E400001-B5A3-F393-E0A9-E50E24DCCA9E
// - RX:      6E400002-B5A3-F393-E0A9-E50E24DCCA9E (write)
// - TX:      6E400003-B5A3-F393-E0A9-E50E24DCCA9E (notify)
//
// Note: This is a "central" implementation (connects to peripherals advertising NUS).
// Peripheral mode (advertising NUS) can be added later for phone-to-phone.
type iosTransport struct {
	h C.reticulum_ble_t
}

func New(target string) (Transport, error) {
	name, addr := ParseTarget(target)
	cName := C.CString(name)
	cAddr := C.CString(addr)
	h := C.reticulum_ble_new(cName, cAddr)
	C.free(unsafe.Pointer(cName))
	C.free(unsafe.Pointer(cAddr))
	if h == nil {
		return nil, errors.New("ble: failed to create CoreBluetooth transport")
	}
	return &iosTransport{h: h}, nil
}

func (t *iosTransport) String() string {
	if t == nil || t.h == nil {
		return "ble://"
	}
	if s := C.reticulum_ble_debug_string(t.h); s != nil {
		return C.GoString(s)
	}
	return "ble://"
}

func (t *iosTransport) Open(ctx context.Context) error {
	if t == nil || t.h == nil {
		return errors.New("ble: nil transport")
	}
	timeout := 7.0
	if deadline, ok := ctx.Deadline(); ok {
		_ = deadline
		// Keep a simple fixed timeout for now; CoreBluetooth callbacks are the true driver.
	}
	if rc := C.reticulum_ble_open(t.h, C.double(timeout)); rc != 0 {
		return errors.New("ble: open failed")
	}
	return nil
}

func (t *iosTransport) Close() error {
	if t == nil || t.h == nil {
		return nil
	}
	C.reticulum_ble_close(t.h)
	C.reticulum_ble_free(t.h)
	t.h = nil
	return nil
}

func (t *iosTransport) IsOpen() bool {
	if t == nil || t.h == nil {
		return false
	}
	return C.reticulum_ble_is_open(t.h) != 0
}

func (t *iosTransport) DeviceDisappeared() bool {
	if t == nil || t.h == nil {
		return false
	}
	return C.reticulum_ble_device_disappeared(t.h) != 0
}

func (t *iosTransport) Read(p []byte) (int, error) {
	if t == nil || t.h == nil {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	n := C.reticulum_ble_read(t.h, (*C.uint8_t)(unsafe.Pointer(&p[0])), C.int(len(p)))
	if n < 0 {
		return 0, io.EOF
	}
	return int(n), nil
}

func (t *iosTransport) Write(p []byte) (int, error) {
	if t == nil || t.h == nil {
		return 0, io.EOF
	}
	if len(p) == 0 {
		return 0, nil
	}
	n := C.reticulum_ble_write(t.h, (*C.uint8_t)(unsafe.Pointer(&p[0])), C.int(len(p)))
	if n < 0 {
		return 0, io.EOF
	}
	return int(n), nil
}
