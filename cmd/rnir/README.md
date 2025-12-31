# rnir

Reticulum Identity Resolver utility (Go port). This tool currently acts as a lightweight Reticulum initialiser and parity stub.

## Overview

- Initialises Reticulum with the selected config and logging settings.
- Provides `--service` mode to log to file.
- Prints a small example config snippet (`--exampleconfig`).

## Usage

`rnir [--config <dir>] [-v|-q ...] [--service]`

## Flags

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-config` | string | `""` | Alternative Reticulum config directory (defaults to `~/.reticulum`). |
| `-v`, `-verbose` | count | `0` | Increase verbosity. |
| `-q`, `-quiet` | count | `0` | Decrease verbosity. |
| `-service` | bool | `false` | Log to `logfile` and disable verbosity output. |
| `-exampleconfig` | bool | `false` | Print an example config snippet and exit. |
| `-version` | bool | `false` | Print version and exit. |

## Examples

Initialise using default config:

`go run ./cmd/rnir`

Service mode:

`go run ./cmd/rnir -config "$HOME/.reticulum-go" -service`

## Exit codes

- `0`: Success.
- `1`: Could not start Reticulum.

