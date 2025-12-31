# rncp

Reticulum file transfer utility for the Go port. It can send files to a listener, receive files, or fetch files from a remote listener.

## Overview

- Listener mode (`-l`) accepts incoming file transfers (and optionally fetch requests).
- Sender mode sends a local file to a destination hash.
- Fetch mode (`-f`) requests a file from a remote listener (requires `-F` on the listener).
- Supports simple access control lists (`-a`) or unauthenticated mode (`-n`).

## Usage

Send a file:

`rncp [options] <file> <destination_hash>`

Listen for incoming transfers:

`rncp --listen [options]`

Fetch a file from a remote listener:

`rncp --fetch [options] <remote_path> <destination_hash>`

## Flags

| Flag(s) | Type | Default | Description |
|---|---:|---:|---|
| `-config` | string | `""` | Alternative Reticulum config directory (defaults to `~/.reticulum`). |
| `-i`, `-identity` | string | `""` | Identity file path to use (listener and client). |
| `-l`, `-listen` | bool | `false` | Listen for incoming transfers. |
| `-p`, `-print-identity` | bool | `false` | Print identity and listening destination hash and exit. |
| `-n`, `-no-auth` | bool | `false` | Accept requests from anyone (disables ACL). |
| `-a` | repeatable string | `[]` | Allow this identity hash (in hex). |
| `-b` | int | `-1` | Announce interval in seconds (`0` = announce once at startup, `-1` = do not announce). |
| `-w` | float | `15` | Sender timeout before giving up (seconds). |
| `-C`, `-no-compress` | bool | `false` | Disable automatic compression. |
| `-S`, `-silent` | bool | `false` | Disable transfer progress output. |
| `-P`, `-phy-rates` | bool | `false` | Display physical layer transfer rates (when available). |
| `-s`, `-save` | string | `""` | Save received files in this directory. |
| `-O`, `-overwrite` | bool | `false` | Allow overwriting received files (instead of auto postfix). |
| `-f`, `-fetch` | bool | `false` | Fetch file from remote listener instead of sending. |
| `-F`, `-allow-fetch` | bool | `false` | Allow authenticated clients to fetch files (listener-side). |
| `-j`, `-jail` | string | `""` | Restrict fetch requests to paths under this directory (listener-side). |
| `-v`, `-verbose` | count | `0` | Increase verbosity. |
| `-q`, `-quiet` | count | `0` | Decrease verbosity. |
| `-h`, `-help` | bool | `false` | Print help and exit. |
| `-version` | bool | `false` | Print version and exit. |

## Examples

Print the listening destination hash:

`go run ./cmd/rncp -p`

Start a permissive listener (no ACL) and save files:

`go run ./cmd/rncp -l -n -s "$HOME/Downloads/rncp"`

Send a file:

`go run ./cmd/rncp ./hello.txt <RNCP_DEST_HASH>`

Fetch a file (remote must be started with `-F`):

`go run ./cmd/rncp -f /remote/path/file.txt <RNCP_DEST_HASH>`

## Exit codes

- `0`: Success (or help/version printed).
- `1`: Generic failure.
- `2`: Usage error (not enough arguments).

Additional codes may be used for specific file/permission errors.

