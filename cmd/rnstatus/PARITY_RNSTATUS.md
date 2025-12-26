# rnstatus parity TODO

Only items still outstanding vs `python/RNS/Utilities/rnstatus.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Decide on identity recall semantics in remote mode: Go falls back to `IdentityRecall(..., fromIdentityHash=true)`; Python uses only `Identity.recall(destination_hash)`. If strict parity is required, remove/guard the fallback and match Python’s error text.
- Remote request failure code-path: Python prints “The remote status request failed…” and then caller exits with code `2`; Go prints the message but returns `nil` stats (caller currently maps to code `2`), verify all branches match exactly (including when link closes vs auth failure).
- JSON output: Python `json.dumps(stats)` ordering/spacing differs from Go `json.Marshal`; if 1:1 output is needed, add Python-like formatting (as done in `rnpath`), otherwise document difference.
- Help/argparse parity: Go emulates argparse-style `usage:` and “unrecognized arguments” errors; verify edge cases (unknown flag formatting, `filter` positional, exit code `2`) match Python.

