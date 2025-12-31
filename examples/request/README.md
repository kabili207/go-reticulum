# request

## Overview
The `request` example demonstrates request/response handlers over a Link.

It has two modes:
- **Server** registers a handler for `/random/text` and returns a random string.
- **Client** connects to the server and triggers requests interactively.

## Usage
### 1) Start the server
```bash
go run ./cmd/rnsd
go run ./examples/request -server
```

Copy the printed destination hash like `<…>`.

### 2) Run the client
```bash
go run ./examples/request -destination <destination_hash_hex>
```

Press Enter in the client to send a request.

## Flags
| Flag | Description |
| --- | --- |
| `-server` | Run in server mode (wait for incoming requests). |
| `-destination <hex>` | Server destination hash (client mode). |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

