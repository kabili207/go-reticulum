# rnsd

Reticulum Network Stack daemon for the Go port. It starts (or connects to) a local shared instance and keeps it running.

## Overview

- Loads a Reticulum configuration and brings up configured interfaces.
- Creates a shared instance by default so other utilities can talk to it (via LocalInterface IPC).
- Supports service mode (log to file) and an interactive Go REPL for debugging.

## Interfaces

`rnsd` reads the `[interfaces]` section from the Reticulum config file and brings up each enabled interface. The Go port focuses on parity with the Python reference configuration format, but not every Python interface is fully implemented yet.

### Common interface keys

Most interfaces accept a shared set of keys (names may be aliases in configs):

- `enabled` / `interface_enabled`: must be truthy (`Yes/True/1`) for the interface to be started.
- `name`: a human-friendly name (usually the section name is used).
- `selected_interface_mode` / `interface_mode` / `mode`: interface operating mode (`full`, `ap`, `ptp`, `roaming`, `boundary`, `gateway`).
- `configured_bitrate` / `bitrate`: sets the bitrate used for scheduling/MTU heuristics.
- `ifac_size`, `networkname`, `passphrase`: optional IFAC settings (interface authentication).
- `announce_cap`, `announce_rate_target`, `announce_rate_grace`, `announce_rate_penalty`: announce rate limiting.

### Supported interface types

The `type` key selects the driver. The Go port supports these internal types:

- `AutoInterface` — IPv6 link-local peer discovery (useful on LANs). Minimal config, but may produce OS-level warnings on constrained interfaces.
- `UDPInterface` — UDP broadcast/unicast interface. Typical keys: `listen_ip`, `listen_port`, `forward_ip`, `forward_port`, `device`.
- `TCPClientInterface` — outbound TCP uplink to a remote instance. Keys: `target_host`, `target_port` (or `port`), optional `kiss_framing`, `connect_timeout`, `reconnect_wait`, `max_reconnect_tries`.
- `TCPServerInterface` — inbound TCP listener that spawns per-client sub-interfaces. Keys: `listen_ip` (or `device`), `listen_port` (or `port`), optional `prefer_ipv6`, `kiss_framing`.
- `KISSInterface` — serial KISS TNC. Keys: `port`, `speed`, optional framing/flow-control parameters.
- `AX25KISSInterface` — AX.25 over KISS.
- `SerialInterface` — raw serial interface (non-KISS).
- `RNodeInterface` / `RNodeMultiInterface` — LoRa RNode devices.
- `I2PInterface` — I2P tunneled interface (requires I2P environment).
- `WeaveInterface` — Weave interface (uses a persisted per-port identity).
- `PipeInterface` — pipe-based interface for local process integration.
- `BackboneInterface` / `BackboneClientInterface` — backbone transport links (client/server selection depends on config).

### Not supported (Python external interfaces)

Python Reticulum can load external interface modules from `interfacepath` as `<Type>.py`. The Go port does not execute Python modules. If a matching external module is found, startup errors instead of silently running without it.

## Usage

Start the daemon:

`rnsd [--config <dir>] [-v|-q ...] [--service] [--interactive]`

Print the verbose example configuration (and exit):

`rnsd --exampleconfig`

## Flags

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-config` | string | `""` | Use an alternative Reticulum config directory (defaults to `~/.reticulum`). |
| `-v`, `-verbose` | count | `0` | Increase verbosity (repeatable). |
| `-q`, `-quiet` | count | `0` | Decrease verbosity (repeatable). |
| `-s`, `-service` | bool | `false` | Run as a service and log to file instead of stdout. |
| `-i`, `-interactive` | bool | `false` | Drop into an interactive Go shell after startup. |
| `-exampleconfig` | bool | `false` | Print a verbose example config and exit. |
| `-version` | bool | `false` | Print version and exit. |
| `-h`, `-help` | bool | `false` | Print help and exit. |

## Examples

Start using the default config (`~/.reticulum/config`):

`go run ./cmd/rnsd`

Start with a specific config directory:

`go run ./cmd/rnsd -config "$HOME/.reticulum-go"`

Service mode (write `logfile` inside the config directory):

`go run ./cmd/rnsd -config "$HOME/.reticulum-go" -service`

Interactive mode (provides `rnsd.Ret` as the Reticulum instance):

`go run ./cmd/rnsd -interactive`

## Exit codes

- `0`: Success (or help/example/version printed).
- `1`: Startup error.
- `2`: CLI usage / unknown flag / unexpected positional arguments.
