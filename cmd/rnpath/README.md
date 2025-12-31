# rnpath

Reticulum path discovery and path table management utility for the Go port.

## Overview

- Requests discovery of a path to a destination hash.
- Prints the local path table and filters by hop count.
- Drops individual paths, queued announces, or paths via a given transport instance.
- Can operate against a remote transport instance via Remote Management (`-R` + `-i`).

## Usage

Discover a path:

`rnpath [options] <destination_hash>`

Show the local path table:

`rnpath -t [options]`

Remote management mode:

`rnpath -R <transport_identity_hash> -i <identity_path> [options]`

## Flags

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-config` | string | `""` | Alternative Reticulum config directory (defaults to `~/.reticulum`). |
| `-t`, `-table` | bool | `false` | Show all known paths. |
| `-m`, `-max` | int | `-1` | Filter path table by maximum hops (`-1` disables). |
| `-r`, `-rates` | bool | `false` | Show announce rate info. |
| `-d`, `-drop` | bool | `false` | Remove the path to a destination (requires `<destination_hash>`). |
| `-D`, `-drop-announces` | bool | `false` | Drop all queued announces. |
| `-x`, `-drop-via` | bool | `false` | Drop all paths via specified transport instance (requires `<destination_hash>` as the via hash). |
| `-w` | float | `15` | Timeout before giving up (seconds). |
| `-R` | string | `""` | Remote mode: transport identity hash of remote instance to manage. |
| `-i` | string | `""` | Remote mode: management identity path. |
| `-W` | float | `15` | Remote mode: timeout for remote queries (seconds). |
| `-j`, `-json` | bool | `false` | Output in JSON format (where applicable). |
| `-v`, `-verbose` | count | `0` | Increase verbosity. |
| `-V`, `-version` | bool | `false` | Print version and exit. |

## Examples

Discover a path:

`go run ./cmd/rnpath 64a2620223471e626954c03d514e674d`

Show full table:

`go run ./cmd/rnpath -t`

Show table filtered to <= 3 hops:

`go run ./cmd/rnpath -t -m 3`

Drop one destination path:

`go run ./cmd/rnpath -d 64a2620223471e626954c03d514e674d`

Drop queued announces:

`go run ./cmd/rnpath -D`

## Exit codes

- `0`: Success.
- `1`: General error.
- `12`: Path request timed out (remote management connect).
- `20`: Remote management errors.

