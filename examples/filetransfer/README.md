# filetransfer

## Overview
The `filetransfer` example demonstrates:
- Request/response handlers (`/files`, `/get`)
- Link establishment
- Sending files as `Resource` transfers

It has two modes:
- **Server** serves a directory of files to clients.
- **Client** lists server files and downloads selected files into an output directory.

## Usage
### 1) Start the server
```bash
go run ./cmd/rnsd
go run ./examples/filetransfer -serve ./some-dir
```

Copy the printed server destination hash like `<…>`.

### 2) Run the client
```bash
go run ./examples/filetransfer -destination <destination_hash_hex> -out ./downloads
```

## Flags
| Flag | Description |
| --- | --- |
| `-serve <dir>` | Serve a directory of files (server mode). |
| `-destination <hex>` | Server destination hash (client mode). |
| `-out <dir>` | Output directory for downloaded files (client mode). |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

## Notes
- The server validates file paths and only serves regular files from the served directory.
- Transfers run over an established link; the server starts a `Resource` when `/get` is requested.

