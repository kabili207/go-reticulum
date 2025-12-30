# rnsd parity TODO

Only items still outstanding vs `python/RNS/Utilities/rnsd.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Interactive mode parity: Python uses `code.interact(local=globals())`; Go uses Yaegi and currently only exposes `rnsd.Ret` (not `RNS`, module globals, helpers, etc.). Decide the desired interactive surface area and align UX (prompt/banner/exit commands).
- Help/usage output: Python `argparse` formatting vs Go `flag` output will differ; decide if strict 1:1 help text is needed.

## Covered by integration tests

- `tests/integration/compare_rnsd_py_vs_go.sh` covers `--version`, `--exampleconfig`, `--service` logfile behaviour, rejecting unknown flags like `--quickchecks`, and basic shared-instance/rpc_key scenarios validated via `rnstatus`.
