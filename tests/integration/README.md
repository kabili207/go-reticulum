# Integration scripts

Shell wrappers for running the `-tags=integration` Go test suites for the CLI binaries under `cmd/`.

The old locations under `tests/run_*_integration.sh` are kept as tiny shims for backwards compatibility.

## Extra regression helpers

- `tests/run_all_parity.sh` — runs Go tests, Python tests, then all parity scripts under `tests/integration/` in a stable sequential order.
- `tests/integration/compare_rnstatus_py_vs_go.sh` — runs `rnsd` (Python and Go) for each template config under `configs/testing/**/config`, fetches `rnstatus -j -a`, normalizes output and diffs it, plus compares text-mode output across a small flag matrix.
  - Requires: `python3`, Go toolchain.
  - Writes logs and diffs to `tests/_logs/<timestamp>/compare_rnstatus/`.
  - Env:
    - `SHARED_INSTANCE_TYPE=unix` — forces `shared_instance_type` in test configs (useful in environments where loopback TCP is not permitted).
- `tests/integration/compare_rnsd_py_vs_go.sh` — compares `rnsd` flag behaviour (Python vs Go) and validates shared-instance/rpc_key basics via `rnstatus`.
  - Requires: `python3`, Go toolchain.
  - Writes logs and diffs to `tests/_logs/<timestamp>/compare_rnsd/`.
- `tests/integration/compare_rnid_py_vs_go.sh` — compares `rnid` identity/crypto workflows (Python vs Go) without network by cross-validating identity load, export/import, sign/validate, and encrypt/decrypt.
  - Requires: `python3`, Go toolchain.
  - Writes logs and diffs to `tests/_logs/<timestamp>/compare_rnid/`.
- `tests/integration/compare_rnpath_py_vs_go.sh` — compares `rnpath` CLI behaviour (Python vs Go) for `-t/-m/-r/-d/-x/-D/-j`, including exit codes and output (with Reticulum log lines stripped for stable diffs).
  - Requires: `python3`, Go toolchain.
  - Writes logs and diffs to `tests/_logs/<timestamp>/compare_rnpath/`.
- `tests/integration/compare_rnprobe_py_vs_go.sh` — compares `rnprobe` CLI behaviour (Python vs Go) for argument validation and `-t/-v/--version` with stable output normalization (spinner/control chars removed).
  - Requires: `python3`, Go toolchain.
  - Writes logs and diffs to `tests/_logs/<timestamp>/compare_rnprobe/`.
- `tests/integration/compare_rnprobe_two_nodes_py_vs_go.sh` — starts two local `rnsd` nodes (Python then Go) using `configs/testing/two_nodes_udp/**`, extracts the probe responder hash from `rnstatus`, and verifies `rnprobe rnstransport.probe <hash>` succeeds with 0% loss (summary parity).
  - Requires: `python3`, Go toolchain, ability to bind local UDP sockets.
  - Writes logs and summaries to `tests/_logs/<timestamp>/compare_rnprobe_two_nodes/`.
- `tests/integration/compare_rnpath_two_nodes_py_vs_go.sh` — starts two local `rnsd` nodes (Python then Go) using `configs/testing/two_nodes_udp/**`, discovers a path to the probe responder hash, verifies `-t` and `-t -j` output, and checks `-d` matches Python’s expire-path semantics (summary parity).
  - Requires: `python3`, Go toolchain, ability to bind local UDP sockets.
  - Writes logs and summaries to `tests/_logs/<timestamp>/compare_rnpath_two_nodes/`.
- `tests/integration/compare_rnir_py_vs_go.sh` — compares `rnir` basics (Python vs Go) for `--version`, `--exampleconfig`, and a real startup using a UDP interface config (bind sockets).
  - Requires: `python3`, Go toolchain, ability to bind local UDP sockets.
  - Writes logs and diffs to `tests/_logs/<timestamp>/compare_rnir/`.
- `tests/integration/compare_rncp_two_nodes_py_vs_go.sh` — starts two local `rnsd` nodes (Python then Go) using `configs/testing/two_nodes_udp/**`, then tests `rncp` send (A→B) and fetch (A←B with `-F` + `-j`) by verifying file hashes (summary parity).
  - Requires: `python3`, Go toolchain, ability to bind local UDP sockets.
  - Writes logs and summaries to `tests/_logs/<timestamp>/compare_rncp_two_nodes/`.
- `tests/integration/compare_rncp_py_vs_go.sh` — compares `rncp` CLI behaviour (Python vs Go) offline: `--version`, `-h`, destination validation, `-p` identity printing via a shared identity file (`-i`), and basic exit-code parity for common error paths.
  - Requires: `python3`, Go toolchain.
  - Writes logs and diffs to `tests/_logs/<timestamp>/compare_rncp/`.
- `tests/integration/compare_rnx_py_vs_go.sh` — compares `rnx` CLI behaviour (Python vs Go) offline: `--version`, `-h`, destination validation and exit codes (including `Path not found` timeout), and `-p` identity printing via a shared identity file (`-i`).
  - Requires: `python3`, Go toolchain.
  - Writes logs and diffs to `tests/_logs/<timestamp>/compare_rnx/`.
- `tests/integration/compare_rnx_two_nodes_py_vs_go.sh` — starts two local `rnsd` nodes (Python then Go) using `configs/testing/two_nodes_udp/**`, then exercises `rnx` listener/client mode: `-n` vs `-a`, `-N`, `-b` (no-announce), `--stdout/--stderr` limits and `-m` mirror exit code (summary parity).
  - Requires: `python3`, Go toolchain, ability to bind local UDP sockets.
  - Writes logs and summaries to `tests/_logs/<timestamp>/compare_rnx_two_nodes/`.
