# resource

## Overview
The `resource` example demonstrates transferring large payloads over a Link using `Resource`.

It has two modes:
- **Server** accepts incoming resources and prints metadata and the received file location.
- **Client** sends a randomly generated resource with metadata.

## Usage
### 1) Start the server
```bash
go run ./cmd/rnsd
go run ./examples/resource -server
```

Copy the printed destination hash like `<…>`.

### 2) Run the client
```bash
go run ./examples/resource -destination <destination_hash_hex> -size-mb 8
```

Press Enter in the client to send a resource.

## Flags
| Flag | Description |
| --- | --- |
| `-server` | Run in server mode (wait for incoming resources). |
| `-destination <hex>` | Server destination hash (client mode). |
| `-size-mb <int>` | Resource size in megabytes (client mode, default `32`). |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

## Notes
- The server logs where the received data is stored (`res.DataFile()`).
- The client sets `auto_compress=false` to match the reference example behaviour.

