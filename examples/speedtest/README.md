# speedtest

## Overview
The `speedtest` example measures link throughput by sending packets until a data cap is reached, then printing transfer statistics.

It has two modes:
- **Server** receives packets and prints stats after `cap-mb` is reached, then tears down the link.
- **Client** sends packets as fast as possible until `cap-mb` is reached (plus a small buffer).

## Usage
### 1) Start the server
```bash
go run ./cmd/rnsd
go run ./examples/speedtest -server
```

Copy the printed destination hash like `<…>`.

### 2) Run the client
```bash
go run ./examples/speedtest -destination <destination_hash_hex> -cap-mb 2
```

## Flags
| Flag | Description |
| --- | --- |
| `-server` | Run in server mode (wait for incoming clients). |
| `-destination <hex>` | Server destination hash (client mode). |
| `-cap-mb <int>` | Data cap in MiB before printing stats (default `2`). |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

