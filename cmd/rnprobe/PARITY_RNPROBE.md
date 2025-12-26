# rnprobe parity TODO

Only items still outstanding vs `python/RNS/Utilities/rnprobe.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Match positional-args behaviour: Python allows calling with only `<destination_hash>` (then prints “full destination name … must be specified”), while Go currently requires 2 args and prints usage instead.
- Align identity recall semantics: Go falls back to `IdentityRecall(..., fromIdentityHash=true)`; Python uses only `Identity.recall(destination_hash)`. Decide if this should stay as a Go-only robustness tweak.
- Decide on `--probes 0` behaviour: Go clamps `probes<=0` to `1`; Python would currently error (division by zero) after sending none. If strict parity is desired, replicate Python’s behaviour (or upstream-fix and track that).
- Verify user-facing strings exactly match (spinner/status lines, MTU error uses raw size vs packed size, whitespace/newlines).

