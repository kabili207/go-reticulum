# rnid parity TODO

Only items still outstanding vs `python/RNS/Utilities/rnid.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Match Python CLI help visibility: Python suppresses `--stdin/--stdout` in `argparse` help; Go currently shows them via `flag` defaults.
- Align a few user-facing log strings 1:1 (eg. “Received Identity … from the network” and spinner/status lines).
- Confirm base64 formatting expectations: Python uses `base64.urlsafe_b64encode()` (padded), Go uses `base64.URLEncoding.EncodeToString()` (also padded) but decoding accepts both padded/raw; verify this is acceptable for strict parity.
- Real-world verification: test `-R/--request` identity fetch timing and `--timeout` behaviour on real networks/interfaces.

