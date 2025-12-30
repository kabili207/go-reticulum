# Integration scripts

Shell wrappers for running the `-tags=integration` Go test suites for the CLI binaries under `cmd/`.

The old locations under `tests/run_*_integration.sh` are kept as tiny shims for backwards compatibility.

## Extra regression helpers

- `tests/integration/compare_rnstatus_py_vs_go.sh` — runs `rnsd` (Python and Go) for each template config under `configs/testing/**/config`, fetches `rnstatus -j -a`, normalizes output and diffs it.
  - Requires: `python3`, Go toolchain.
  - Writes logs and diffs to `tests/_logs/<timestamp>/compare_rnstatus/`.
  - Env:
    - `SHARED_INSTANCE_TYPE=unix` — forces `shared_instance_type` in test configs (useful in environments where loopback TCP is not permitted).
