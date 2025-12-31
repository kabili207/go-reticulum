# ratchets

## Overview
The `ratchets` example demonstrates destination ratchets and proof-based echo requests.

It has two modes:
- **Server** enables ratchets for its destination and sends proofs for received packets.
- **Client** sends echo requests and prints RTT + signal stats when the proof arrives.

## Usage
### 1) Start the server
```bash
go run ./cmd/rnsd
go run ./examples/ratchets -server
```

Copy the printed destination hash like `<…>`.

### 2) Run the client
```bash
go run ./examples/ratchets -destination <destination_hash_hex> -timeout 10
```

## Flags
| Flag | Description |
| --- | --- |
| `-server` | Run in server mode (wait for incoming packets). |
| `-destination <hex>` | Server destination hash (client mode). |
| `-timeout <seconds>` | Client-side reply timeout (seconds). |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

## Notes
- Ratchets are stored in the OS temp directory as `<dest_hash>.ratchets`.
- The example configures a very small ratchet interval to make rotation observable.

