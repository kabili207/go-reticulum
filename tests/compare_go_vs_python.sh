#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="${LOG_DIR:-"$ROOT/tests/_logs"}"
TS="${TS:-"$(date +"%Y%m%d-%H%M%S")"}"
OUT_DIR="$LOG_DIR/$TS"
mkdir -p "$OUT_DIR"

echo "[cmp] out=$OUT_DIR"
echo "[cmp] RUN_SLOW_TESTS=${RUN_SLOW_TESTS:-}"
echo

# Run python first (spawns subprocess targets inside python/tests/link.py) then go.
"$ROOT/tests/run_python_tests.sh" LOG_DIR="$LOG_DIR" >/dev/null
"$ROOT/tests/run_go_tests.sh" LOG_DIR="$LOG_DIR" >/dev/null

PY_LOG="$(ls -1dt "$LOG_DIR"/*/python/output.log | head -n 1 || true)"
GO_LOG="$(ls -1dt "$LOG_DIR"/*/go/unittest.log | head -n 1 || true)"

if [[ -z "${PY_LOG}" || -z "${GO_LOG}" ]]; then
  echo "[cmp] missing logs (py=$PY_LOG go=$GO_LOG)"
  exit 2
fi

cp -f "$PY_LOG" "$OUT_DIR/python.output.log"
cp -f "$GO_LOG" "$OUT_DIR/go.output.log"

normalize() {
  # Keep PASS/FAIL, test names, and stack traces. Drop noisy perf/timestamp lines.
  # BSD/macOS sed regex quirks make heavy normalization brittle, so prefer filtering.
  grep -Ev \
    -e '^\[[0-9]{4}-[0-9]{2}-[0-9]{2} ' \
    -e 'Mbps|Gbps' \
    -e 'timing min/avg/med/max/mdev' \
    -e 'Max deviation from median' \
    -e '^Sign/validate ' \
    -e '^Testing (random small|large) chunk encrypt/decrypt' \
    -e '^\\s*Encrypt ' \
    -e '^\\s*Decrypt '
}

normalize <"$OUT_DIR/python.output.log" >"$OUT_DIR/python.output.norm.log"
normalize <"$OUT_DIR/go.output.log" >"$OUT_DIR/go.output.norm.log"

echo "[cmp] python log: $OUT_DIR/python.output.log"
echo "[cmp] go log:     $OUT_DIR/go.output.log"
echo

echo "[cmp] python summary:"
grep -E "^(OK$|FAILED \\(|Ran [0-9]+ tests)" "$OUT_DIR/python.output.log" || true
echo

echo "[cmp] go summary:"
grep -E "^(ok\\s|FAIL\\s|\\?\\s)" "$OUT_DIR/go.output.log" || true
echo

echo "[cmp] diff (normalized, first 200 lines):"
diff -u "$OUT_DIR/python.output.norm.log" "$OUT_DIR/go.output.norm.log" | sed -n '1,200p' || true
echo
echo "[cmp] done"
