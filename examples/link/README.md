# link

## Overview
The `link` example demonstrates basic Link establishment and bidirectional packet exchange.

It has two modes:
- **Server** waits for an incoming link and replies to received packets.
- **Client** connects to the server and lets you type messages to send over the link.

## Usage
### 1) Start the server
```bash
go run ./cmd/rnsd
go run ./examples/link -server
```

Copy the printed destination hash like `<…>`.

### 2) Run the client
```bash
go run ./examples/link -destination <destination_hash_hex>
```

## Flags
| Flag | Description |
| --- | --- |
| `-server` | Run in server mode (wait for incoming link requests). |
| `-destination <hex>` | Server destination hash (client mode). |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

## Notes
- The client requests a path and waits for an announce when needed.
- Type `quit` to teardown the link and exit.

