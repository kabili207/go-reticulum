# rnx

Reticulum remote execution utility for the Go port. It can run in listener mode (server) or client mode to execute commands remotely.

## Overview

- Listener mode (`-l`) exposes a destination `rnx.execute` that accepts `command` requests.
- Client mode establishes a link to a listener and requests execution.
- Supports ACL allow-list (`-a`) or open mode (`-n`).
- Supports interactive REPL (`-x`) and per-request timeouts.
- Can limit returned stdout/stderr sizes.

## Usage

Print identity and destination hash (for sharing with clients):

`rnx -p [--config <dir>]`

Start a listener:

`rnx -l [options]`

Execute a command on a listener:

`rnx -destination <listener_hash> -command "<cmd>" [options]`

Interactive client shell:

`rnx -destination <listener_hash> -x [options]`

## Flags

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-config` | string | `""` | Alternative Reticulum config directory (defaults to `~/.reticulum`). |
| `-destination` | string | `""` | Listener destination hash in hex (can also be given as first positional arg). |
| `-command` | string | `""` | Command to execute (can also be given as second positional arg). |
| `-v`, `-verbose` | count | `0` | Increase verbosity (repeatable, supports `-vv`). |
| `-q`, `-quiet` | count | `0` | Decrease verbosity (repeatable, supports `-qq`). |
| `-p`, `-print-identity` | bool | `false` | Print identity and listening destination and exit. |
| `-l`, `-listen` | bool | `false` | Listen for incoming commands. |
| `-i` | string | `""` | Path to identity to use. |
| `-x`, `-interactive` | bool | `false` | Enter interactive client mode. |
| `-b`, `-no-announce` | bool | `false` | Do not announce at program start (listener mode). |
| `-a` | repeatable string | `[]` | Allow commands only from these identity hashes. |
| `-n`, `-noauth` | bool | `false` | Accept commands from anyone (listener mode). |
| `-N`, `-noid` | bool | `false` | Do not identify to listener (client mode). |
| `-d`, `-detailed` | bool | `false` | Print returned stdout/stderr (client mode). |
| `-m` | bool | `false` | Mirror exit code of remote command (non-interactive client mode). |
| `-w` | float | `15` | Connect/request timeout (seconds). |
| `-W` | float | unset | Max result download time (seconds). |
| `-stdin` | string | unset | Pass input to remote stdin. |
| `-stdout` | int | unset | Limit returned stdout bytes. |
| `-stderr` | int | unset | Limit returned stderr bytes. |
| `-version` | bool | `false` | Print version and exit. |
| `-h`, `-help` | bool | `false` | Print help and exit. |

## Examples

Listener (no ACL, announces at startup):

`go run ./cmd/rnx -l -n`

Get listener hash:

`go run ./cmd/rnx -p`

Client execution:

`go run ./cmd/rnx -destination <LISTENER_HASH> -w 15 -command "echo hello" -detailed`

Interactive:

`go run ./cmd/rnx -destination <LISTENER_HASH> -x`

## Exit codes

`rnx` uses explicit exit codes to match upstream behaviour:

- `0`: Success (or interactive mode exited).
- `2`: CLI usage / unknown flag.
- `241`: Invalid destination / Reticulum init / identity load failure.
- `242`: Path not found within timeout.
- `243`: Could not create/establish link with listener.
- `244`: Could not request remote execution.
- `245`: No result received.
- `246`: Receiving result failed.
- `247`: Invalid result payload.
- `248`: Remote could not execute command.
- `249`: No response payload.

