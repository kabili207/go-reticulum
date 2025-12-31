# broadcast

## Overview
The `broadcast` example uses a **plain inbound destination** as a shared broadcast channel. Anything you type is sent as a packet; incoming packets are printed to stdout.

## Usage
Run multiple instances (same `-channel`) to see messages.

```bash
go run ./cmd/rnsd
go run ./examples/broadcast -channel public_information
```

## Flags
| Flag | Description |
| --- | --- |
| `-config <dir>` | Path to an alternative Reticulum config directory. |
| `-channel <name>` | Broadcast channel name (defaults to `public_information`). |

## What to expect
- The program logs its destination hash on startup.
- Type a line and press Enter to broadcast it.
- When another peer broadcasts on the same channel, you will see `Received data: ...`.

