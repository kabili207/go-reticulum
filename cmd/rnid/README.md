# rnid

Reticulum identity and crypto utility for the Go port. It can generate identities, print/export/import them, announce destinations, and encrypt/decrypt/sign/validate files.

## Overview

- Generate identities (`-g`), print (`-p`) and export (`-x`) them.
- Compute destination hashes for different aspects (`-H`).
- Announce a destination name derived from an identity (`-a`).
- File operations (mutually exclusive): `--encrypt`, `--decrypt`, `--sign`, `--validate`.
- Optional network mode: request unknown identities from the network (`-R`) when given a destination hash.

## Usage

`rnid [options]`

The tool is mode-driven by flags; for example:

- Generate: `rnid -g <path>`
- Print: `rnid -i <identity_or_hash_or_path> -p`
- Encrypt/decrypt/sign/validate: `rnid -i <identity_or_hash_or_path> --encrypt ...`

## Flags

### Global

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-config` | string | `""` | Alternative Reticulum config directory (defaults to `~/.reticulum`). |
| `-i`, `-identity` | string | `""` | Identity selector: hex identity hash, destination hash, or path to an identity file. |
| `-v`, `-verbose` | count | `0` | Increase verbosity. |
| `-q`, `-quiet` | count | `0` | Decrease verbosity. |
| `-version` | bool | `false` | Print version and exit. |

### Identity management

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-g`, `-generate` | string | `""` | Generate a new identity and write it to this path. |
| `-m`, `-import` | string | `""` | Import an identity from hex/base32/base64 (use `-w` to write). |
| `-x`, `-export` | bool | `false` | Export the selected identity to stdout (hex/base32/base64). |
| `-p`, `-print-identity` | bool | `false` | Print identity info and exit. |
| `-P`, `-print-private` | bool | `false` | Allow printing private keys (where supported). |
| `-b`, `-base64` | bool | `false` | Use base64 encoding for import/export. |
| `-B`, `-base32` | bool | `false` | Use base32 encoding for import/export. |

### Destination operations

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-a`, `-announce` | string | `""` | Announce a destination based on this identity (format `app.aspect...`). |
| `-H`, `-hash` | string | `""` | Show destination hashes for other aspects (comma/space separated). |

### Network identity lookup

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-R`, `-request` | bool | `false` | Request unknown identities from the network when needed. |
| `-t`, `-timeout` | float | `15` | Timeout for identity requests (seconds). |

### File operations (mutually exclusive)

Only one of these can be used at a time:

| Flag(s) | Type | Description |
|---|---:|---|
| `-e`, `-encrypt` | string | Encrypt file. |
| `-d`, `-decrypt` | string | Decrypt file. |
| `-s`, `-sign` | string | Sign file. |
| `-V`, `-validate` | string | Validate signature (path to `.rsg`). |

And the shared I/O flags:

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-r`, `-read` | string | `""` | Input file path. If omitted, may be derived from `--encrypt/--decrypt/--sign`. |
| `-w`, `-write` | string | `""` | Output file path. |
| `-f`, `-force` | bool | `false` | Overwrite existing output files. |
| `-I`, `-stdin` | bool | `false` | Read input from stdin instead of file. |
| `-O`, `-stdout` | bool | `false` | Write output to stdout instead of file. |

## Examples

Generate an identity:

`go run ./cmd/rnid -g ./my.id`

Print identity:

`go run ./cmd/rnid -i ./my.id -p`

Compute hashes for other aspects:

`go run ./cmd/rnid -i ./my.id -H "app.aspect1 app.aspect2"`

Sign and validate:

`go run ./cmd/rnid -i ./my.id -s ./message.txt -w ./message.rsg -f`

`go run ./cmd/rnid -i ./my.id -V ./message.rsg`

Encrypt and decrypt:

`go run ./cmd/rnid -i ./my.id -e ./secret.txt -w ./secret.rfe -f`

`go run ./cmd/rnid -i ./my.id -d ./secret.rfe -w ./secret.dec -f`

## Exit codes

- `0`: Success (or version printed).
- `2`: No identity provided / usage.
- `3`: Refused to overwrite an existing output file.
- `4`: Identity generation/save error.
- `101`: Could not initialise Reticulum.

