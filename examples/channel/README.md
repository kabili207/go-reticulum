# channel

## Overview
The `channel` example demonstrates message-based communication over a Link `Channel` (MessagePack encoding).

It has two modes:
- **Server mode** waits for a client and replies to `StringMessage` messages.
- **Client mode** connects to the server and sends typed messages over the Channel.

## Usage
### 1) Start the server
```bash
go run ./cmd/rnsd
go run ./examples/channel -server
```

Copy the printed destination hash like `<…>`.

### 2) Run the client
```bash
go run ./examples/channel <destination_hash_hex>
```

## Flags
| Flag | Description |
| --- | --- |
| `-server`, `-s` | Run in server mode (wait for incoming link requests). |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

## Notes
- The client requires a **truncated destination hash** (hex) shown by the server.
- Messages are encoded as `(data, timestamp)`; the timestamp is stored as RFC3339Nano for portability.

