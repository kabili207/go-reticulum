package ble

import (
	"context"
	"strings"
)

type Transport interface {
	Open(ctx context.Context) error
	Close() error
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	IsOpen() bool
	DeviceDisappeared() bool
	String() string
}

func ParseTarget(target string) (name, addr string) {
	t := strings.TrimSpace(target)
	if t == "" {
		return "", ""
	}
	if strings.HasPrefix(strings.ToLower(t), "name:") {
		return strings.TrimSpace(t[len("name:"):]), ""
	}
	if len(t) == 17 && strings.Count(t, ":") == 5 {
		return "", t
	}
	return t, ""
}

func ChunkByMaxWriteNoRsp(b []byte, max int) [][]byte {
	if len(b) == 0 {
		return nil
	}
	if max <= 0 {
		return [][]byte{b}
	}
	if len(b) <= max {
		return [][]byte{b}
	}
	out := make([][]byte, 0, (len(b)+max-1)/max)
	for len(b) > 0 {
		n := max
		if n > len(b) {
			n = len(b)
		}
		out = append(out, b[:n])
		b = b[n:]
	}
	return out
}
