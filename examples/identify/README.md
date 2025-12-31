# identify

## Overview
The `identify` example demonstrates Link identification:
- The client creates an identity and calls `Link.Identify(...)`.
- The server receives the remote identity in `SetRemoteIdentifiedCallback`.

It also exchanges packets over the link after identification.

## Usage
### 1) Start the server
```bash
go run ./cmd/rnsd
go run ./examples/identify -server
```

Copy the printed destination hash like `<…>`.

### 2) Run the client
```bash
go run ./examples/identify -destination <destination_hash_hex>
```

## Flags
| Flag | Description |
| --- | --- |
| `-server` | Run in server mode (wait for incoming link requests). |
| `-destination <hex>` | Server destination hash (client mode). |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

## Notes
- The client will request a path and wait until the server announce arrives (if needed).
- After the link is active and identified, type lines to send packets.

