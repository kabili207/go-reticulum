# rnx parity TODO

Only items still outstanding vs `python/RNS/Utilities/rnx.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Verify command parsing parity: Python uses `shlex.split(command)`; Go uses `splitCommand()`—confirm quoting/escaping and platform differences (Windows vs POSIX) match expectations.
- Confirm timeout/kill semantics on the listen side: Python loops with `process.communicate(timeout=1)` and then `terminate()` after deadline; Go uses `exec.CommandContext` with a context deadline—validate stdout/stderr capture and “concluded” timestamp behaviour under timeouts.
- Align small UX/text details: interactive prompt behaviour, “clear” handling, and any newline/spacing differences in spinner/status messages.
- Optional robustness vs strict parity: Go utilities sometimes add identity-recall fallbacks; ensure `rnx` matches Python where needed (eg. listener identity recall errors and messages).

