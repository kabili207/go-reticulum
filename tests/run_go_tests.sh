#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${LOG_DIR:-"$ROOT/tests/_logs"}"
TS="$(date +"%Y%m%d-%H%M%S")"
OUT_DIR="$LOG_DIR/$TS/go"
mkdir -p "$OUT_DIR"

export RNS_INTEGRATION="${RNS_INTEGRATION:-1}"

echo "[go] root=$ROOT"
echo "[go] out=$OUT_DIR"
echo "[go] RNS_INTEGRATION=$RNS_INTEGRATION"
echo "[go] RUN_SLOW_TESTS=${RUN_SLOW_TESTS:-}"
echo

cd "$ROOT"

set -x
# Raw go output (useful for debugging)
go test ./... -count=1 2>&1 | tee "$OUT_DIR/output.log"

# unittest-like report for parity comparisons
go test ./... -json -count=1 2>/dev/null | python3 "$ROOT/tests/go_unittest_report.py" | tee "$OUT_DIR/unittest.log"
set +x

echo
echo "[go] done"
