# echo

## Overview
The `echo` example demonstrates packet delivery receipts and proofs. The server accepts packets and sends proofs; the client sends "echo requests" and measures RTT when the proof arrives.

## Usage
### 1) Start the server
```bash
go run ./cmd/rnsd
go run ./examples/echo -server
```

Copy the printed destination hash like `<…>`.

### 2) Run the client
```bash
go run ./examples/echo -timeout 10 <destination_hash_hex>
```

Press Enter in the client to send requests. Press Enter in the server to announce.

## Flags
| Flag | Description |
| --- | --- |
| `-server`, `-s` | Run in server mode (wait for incoming packets). |
| `-timeout <seconds>`, `-t <seconds>` | Client-side reply timeout (seconds). |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

## Notes
- The client requests a path if the destination is not yet known.
- RTT is printed when a valid proof is received.

