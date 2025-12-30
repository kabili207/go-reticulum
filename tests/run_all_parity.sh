#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# This script is the "run everything" entrypoint:
# - Go unit + integration tests (includes local sockets)
# - Python unit tests (starts a local RNS instance)
# - Go vs Python parity regression scripts (some start multiple local nodes)
#
# Note: On macOS with sandboxing, you may need escalated permissions for local
# socket operations. If you see "Operation not permitted", re-run with the
# appropriate permission/approval mode.

RUN_NETWORK="${RUN_NETWORK:-1}"

usage() {
  cat <<'EOF'
Usage: tests/run_all_parity.sh [--no-network]

Runs:
  - tests/run_go_tests.sh
  - tests/run_python_tests.sh
  - tests/integration/compare_*_py_vs_go.sh
  - tests/integration/compare_*_two_nodes_py_vs_go.sh (unless --no-network)

Env:
  RUN_NETWORK=1|0  (default 1; --no-network sets 0)
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi
if [[ "${1:-}" == "--no-network" ]]; then
  RUN_NETWORK=0
  shift
fi
if [[ "$#" -ne 0 ]]; then
  usage
  exit 2
fi

echo "[all] root=$ROOT"
echo "[all] RUN_NETWORK=$RUN_NETWORK"

echo
echo "[all] Go tests"
bash "$ROOT/tests/run_go_tests.sh"

echo
echo "[all] Python tests"
bash "$ROOT/tests/run_python_tests.sh"

echo
echo "[all] Parity regression (offline)"
offline_scripts=(
  "$ROOT/tests/integration/compare_rnid_py_vs_go.sh"
  "$ROOT/tests/integration/compare_rnpath_py_vs_go.sh"
  "$ROOT/tests/integration/compare_rnprobe_py_vs_go.sh"
  "$ROOT/tests/integration/compare_rnx_py_vs_go.sh"
  "$ROOT/tests/integration/compare_rncp_py_vs_go.sh"
  "$ROOT/tests/integration/compare_rnsd_py_vs_go.sh"
)
for s in "${offline_scripts[@]}"; do
  echo "[all] $(basename "$s")"
  bash "$s"
done

echo
echo "[all] Parity regression (local instances)"
echo "[all] compare_rnstatus_py_vs_go.sh"
bash "$ROOT/tests/integration/compare_rnstatus_py_vs_go.sh"

if [[ "$RUN_NETWORK" -eq 1 ]]; then
  echo
  echo "[all] Parity regression (two nodes)"
  two_node_scripts=(
    "$ROOT/tests/integration/compare_rnprobe_two_nodes_py_vs_go.sh"
    "$ROOT/tests/integration/compare_rncp_two_nodes_py_vs_go.sh"
    "$ROOT/tests/integration/compare_rnpath_two_nodes_py_vs_go.sh"
    "$ROOT/tests/integration/compare_rnx_two_nodes_py_vs_go.sh"
  )
  for s in "${two_node_scripts[@]}"; do
    echo "[all] $(basename "$s")"
    bash "$s"
  done
else
  echo
  echo "[all] Skipping two-node parity scripts (RUN_NETWORK=0)"
fi

echo
echo "[all] OK"

