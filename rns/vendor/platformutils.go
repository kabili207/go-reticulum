package vendor

import (
	"os"
	"runtime"
	"strings"
)

func GetPlatform() string {
	// как в Python: сначала руками детектим Android по env
	if _, ok := os.LookupEnv("ANDROID_ARGUMENT"); ok {
		return "android"
	}
	if _, ok := os.LookupEnv("ANDROID_ROOT"); ok {
		return "android"
	}

	// дальше просто runtime.GOOS: "linux", "darwin", "windows", ...
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

// PlatformChecks – в Python проверял версию интерпретатора на Windows.
// В Go это особо не нужно, но оставим хук на всякий случай.
func PlatformChecks() {
	if IsWindows() {
		// если надо – можно тут проверять версию Go или окружение
		// сейчас просто no-op
	}
}

// CryptographyOldAPI в Go не имеет смысла, заглушка на будущее.
func CryptographyOldAPI() bool {
	return false
}
