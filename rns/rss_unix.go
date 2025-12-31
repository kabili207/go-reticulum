//go:build !windows

package rns

import (
	"runtime"
	"syscall"
)

func processRSSBytesPlatform() any {
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		return nil
	}
	// Linux reports kilobytes, Darwin reports bytes.
	v := ru.Maxrss
	if runtime.GOOS == "linux" {
		v *= 1024
	}
	if v <= 0 {
		return nil
	}
	return v
}

