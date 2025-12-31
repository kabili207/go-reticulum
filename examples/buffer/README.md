# buffer

## Overview
The `buffer` example demonstrates `rns.Buffer` over a Link channel using a bidirectional buffered read/writer.

It has two modes:
- **Server mode** waits for a client, then echoes received buffer data back to the client.
- **Client mode** connects to the server, sends typed lines over the buffer, and prints the server response.

## Usage
### 1) Start the server
```bash
go run ./cmd/rnsd
go run ./examples/buffer -server
```

The server prints a destination hash like `<…>`. Copy it.

### 2) Run the client
```bash
go run ./examples/buffer <destination_hash_hex>
```

## Flags
| Flag | Description |
| --- | --- |
| `-server`, `-s` | Run in server mode (wait for incoming link requests). |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

## Notes
- The client requires a **truncated destination hash** (hex) shown by the server.
- The client requests a path and waits until the server announce is received.

