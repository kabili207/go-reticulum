# rnir parity TODO

Only items still outstanding vs `python/RNS/Utilities/rnir.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Decide on `--service` parity: Go exposes `--service` (log to file + create `logfile` marker), but Python CLI does not expose service mode (only `program_setup(..., service=False)` exists). Either document as Go-only extension or hide/align behaviour.
- Align failure handling: Python `program_setup()` exits `0` after starting Reticulum; Go returns an error and exits `1` with “Could not start Reticulum…”. If strict parity is desired, match exit codes/messages on startup failures.
- Match help/usage output: Python `argparse` default formatting differs from Go `flag` output (minor UX parity).

