package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	rns "main/rns"
	ifaces "main/rns/interfaces"
)

func main() {
	var (
		ifName    string
		port      string
		speed     int
		databits  int
		parity    string
		stopbits  int
		mode      string
		configDir string
	)

	flag.StringVar(&ifName, "name", "Example Custom Interface", "interface name")
	flag.StringVar(&port, "port", "", "serial port (eg. /dev/ttyUSB0)")
	flag.IntVar(&speed, "speed", 115200, "baudrate")
	flag.IntVar(&databits, "databits", 8, "data bits")
	flag.StringVar(&parity, "parity", "none", "parity: none|even|odd")
	flag.IntVar(&stopbits, "stopbits", 1, "stop bits: 1|2")
	flag.StringVar(&mode, "mode", "gateway", "interface mode: full|point_to_point|access_point|roaming|boundary|gateway")
	flag.StringVar(&configDir, "config", "", "path to alternative Reticulum config directory (optional)")
	flag.Parse()

	if port == "" {
		fmt.Fprintln(os.Stderr, "missing -port")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Reticulum config example (Python version loads custom interfaces dynamically):")
		fmt.Fprintln(os.Stderr, configSnippet(ifName))
		os.Exit(2)
	}

	var configArg *string
	if strings.TrimSpace(configDir) != "" {
		configArg = &configDir
	}
	reticulum, err := rns.NewReticulum(configArg, nil, nil, nil, false, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "NewReticulum:", err)
		os.Exit(1)
	}

	kv := map[string]string{
		"port":     port,
		"speed":    fmt.Sprintf("%d", speed),
		"databits": fmt.Sprintf("%d", databits),
		"parity":   parity,
		"stopbits": fmt.Sprintf("%d", stopbits),
	}
	ifc, err := ifaces.NewSerialInterface(ifName, kv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "NewSerialInterface:", err)
		os.Exit(1)
	}

	reticulum.AddInterface(
		ifc,
		parseMode(mode),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	rns.Log("Interface example running. Press Ctrl-C to stop.", rns.LogInfo)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch
	rns.Log("Stopping.", rns.LogInfo)
}

func configSnippet(name string) string {
	return fmt.Sprintf(`[[%s]]
type = ExampleInterface
enabled = no
mode = gateway
port = /dev/ttyUSB0
speed = 115200
databits = 8
parity = none
stopbits = 1`, name)
}

func parseMode(mode string) int {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "full":
		return ifaces.InterfaceModeFull
	case "point_to_point", "p2p":
		return ifaces.InterfaceModePointToPoint
	case "access_point", "ap":
		return ifaces.InterfaceModeAccessPoint
	case "roaming":
		return ifaces.InterfaceModeRoaming
	case "boundary":
		return ifaces.InterfaceModeBoundary
	case "gateway":
		return ifaces.InterfaceModeGateway
	default:
		return ifaces.InterfaceModeGateway
	}
}
