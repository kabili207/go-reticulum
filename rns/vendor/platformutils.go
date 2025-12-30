package vendor

import (
	"os"
	"runtime"
	"strings"
)

func GetPlatform() string {
	// Like Python: detect Android via environment variables first.
	if _, ok := os.LookupEnv("ANDROID_ARGUMENT"); ok {
		return "android"
	}
	if _, ok := os.LookupEnv("ANDROID_ROOT"); ok {
		return "android"
	}

	// Then fall back to runtime.GOOS: "linux", "darwin", "windows", ...
	return runtime.GOOS
}

func IsLinux() bool {
	return GetPlatform() == "linux"
}

func IsDarwin() bool {
	return GetPlatform() == "darwin"
}

func IsAndroid() bool {
	return GetPlatform() == "android"
}

func IsWindows() bool {
	return strings.HasPrefix(GetPlatform(), "win")
}

func UseEpoll() bool {
	return IsLinux() || IsAndroid()
}

func UseAFUnix() bool {
	// We use filesystem UNIX domain sockets where available to avoid binding TCP ports
	// (useful in constrained/sandboxed environments) and to match upstream behaviour.
	return IsLinux() || IsAndroid() || IsDarwin()
}

// PlatformChecks: Python checked interpreter versions on Windows.
// In Go this is mostly unnecessary, but keep a hook just in case.
func PlatformChecks() {
	if IsWindows() {
		// If needed, check Go runtime or environment here; currently a no-op.
	}
}

// CryptographyOldAPI is not meaningful in Go; kept as a future stub.
func CryptographyOldAPI() bool {
	return false
}
