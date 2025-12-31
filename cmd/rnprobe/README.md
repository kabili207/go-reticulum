# rnprobe

Reticulum probe utility for the Go port. It sends packets to a destination and measures delivery / RTT based on receipts.

## Overview

- Sends `N` probe packets with a random payload to a destination.
- Waits for delivery receipts and prints RTT and hop count.
- Useful when the destination is configured to return proofs (for example, `rnstransport.probe` when `respond_to_probes = True`).

## Usage

`rnprobe [options] <full_name> <destination_hash>`

Where:

- `full_name` is a destination name like `rnstransport.probe` or `rnx.execute`.
- `destination_hash` is a truncated destination hash in hex (32 hex characters by default).

## Flags

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-config` | string | `""` | Alternative Reticulum config directory (defaults to `~/.reticulum`). |
| `-s`, `-size` | int | `16` | Probe payload size (bytes). |
| `-n`, `-probes` | int | `1` | Number of probes to send. |
| `-t`, `-timeout` | float | auto | Timeout before giving up (seconds). When unset, uses `DEFAULT_TIMEOUT + first_hop_timeout`. |
| `-w`, `-wait` | float | `0` | Wait between probes (seconds). |
| `-v`, `-verbose` | count | `0` | Increase verbosity. |
| `-version` | bool | `false` | Print version and exit. |
| `-h`, `-help` | bool | `false` | Print help and exit. |

## Examples

Probe the local transport probe responder:

1) Enable responding to probes in your config:

`respond_to_probes = True`

2) Find the probe responder destination hash:

`go run ./cmd/rnstatus -a`

Look for: `Probe responder at <...> active`

3) Probe it:

`go run ./cmd/rnprobe -n 5 rnstransport.probe <PROBE_DEST_HASH>`

## Exit codes

- `0`: Success (or help/version/usage printed).
- `1`: Path request timed out or a generic failure.
- `2`: Non-zero packet loss (at least one probe timed out).
- `3`: Probe payload exceeds MTU.

