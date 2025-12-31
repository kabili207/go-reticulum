# minimal

## Overview
The `minimal` example is a small end-to-end program that:
- initialises Reticulum
- creates an inbound destination
- sends an announce each time you press Enter

## Usage
```bash
go run ./cmd/rnsd
go run ./examples/minimal
```

## Flags
| Flag | Description |
| --- | --- |
| `-config <dir>` | Path to an alternative Reticulum config directory. |

## Notes
- This example is useful as a quick sanity check that Reticulum initialisation works.

