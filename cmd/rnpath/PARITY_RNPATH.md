# rnpath parity TODO

Only items still outstanding vs `python/RNS/Utilities/rnpath.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Decide whether to keep Go-only CLI extras: `-V` as alias for `--version` (Python only has `--version` via `argparse`).
- Align remote-identity recall behaviour: Go falls back to `IdentityRecall(..., fromIdentityHash=true)` for the remote manager link; Python only does `Identity.recall(destination_hash)`. If strict parity is desired, match Python behaviour and error messages.
- Confirm JSON output formatting: Go normalizes bytes→hex and then applies `pythonJSONSpacing()` to match Python `json.dumps()` output; verify exact whitespace/ordering matches upstream in real runs.
- Real-world verification: remote management mode (`-R/-i/-W`) teardown reasons/messages and timeouts on actual transport instances.

