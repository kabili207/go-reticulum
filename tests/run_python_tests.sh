#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${LOG_DIR:-"$ROOT/tests/_logs"}"
TS="$(date +"%Y%m%d-%H%M%S")"
OUT_DIR="$LOG_DIR/$TS/python"
mkdir -p "$OUT_DIR"

PYTHON="${PYTHON:-python3}"

# The python reticulum sources and upstream python test suite live in "$ROOT/python".
export PYTHONPATH="${PYTHONPATH:-"$ROOT/python"}"

# Ensure subprocess calls to "python" inside the upstream test suite work.
export PATH="$ROOT/tests/_bin:$PATH"

echo "[py] root=$ROOT"
echo "[py] out=$OUT_DIR"
echo "[py] PYTHON=$PYTHON"
echo "[py] PYTHONPATH=$PYTHONPATH"
echo "[py] PATH(prepended)=$ROOT/tests/_bin"
echo "[py] RUN_SLOW_TESTS=${RUN_SLOW_TESTS:-}"
echo "[py] SKIP_NORMAL_TESTS=${SKIP_NORMAL_TESTS:-}"
echo

cd "$ROOT/python"

set -x
# Run upstream python test aggregation.
$PYTHON -m unittest -v tests.all 2>&1 | tee "$OUT_DIR/output.log"
set +x

echo
echo "[py] done"
