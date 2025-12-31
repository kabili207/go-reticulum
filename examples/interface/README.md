# interface

## Overview
The `interface` example demonstrates how to add a custom interface programmatically at runtime (serial interface).

This is the Go equivalent of what the Python version can do via dynamic interface modules, but implemented directly in Go.

## Usage
You must provide a serial port.

```bash
go run ./cmd/rnsd
go run ./examples/interface -port /dev/ttyUSB0
```

## Flags
| Flag | Description |
| --- | --- |
| `-name <string>` | Interface name. |
| `-port <path>` | Serial port (required). |
| `-speed <int>` | Baudrate (default `115200`). |
| `-databits <int>` | Data bits (default `8`). |
| `-parity <none|even|odd>` | Parity setting (default `none`). |
| `-stopbits <1|2>` | Stop bits (default `1`). |
| `-mode <mode>` | Interface mode: `full|point_to_point|access_point|roaming|boundary|gateway` (default `gateway`). |
| `-config <dir>` | Path to an alternative Reticulum config directory (optional). |

## Notes
- If `-port` is missing, the program prints a Reticulum config snippet and exits with code `2`.
- This example waits until interrupted (Ctrl-C).

