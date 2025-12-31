# announce

## Overview
The `announce` example creates two inbound destinations and periodically sends announces with app data. It also registers an announce handler and prints received announces that match an aspect filter.

## Usage
This example is interactive: press Enter to send announces.

```bash
go run ./cmd/rnsd
go run ./examples/announce
```

## Flags
| Flag | Description |
| --- | --- |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

## What to expect
- On each Enter press it sends two announces (fruits + noble gases) and logs the destination hashes.
- If another node announces `example_utilities.announcesample.fruits`, it will be printed by the handler.

