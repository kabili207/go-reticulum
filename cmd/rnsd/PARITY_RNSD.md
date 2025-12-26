# rnsd parity TODO

Only items still outstanding vs `python/RNS/Utilities/rnsd.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Interactive mode parity: Python uses `code.interact(local=globals())`; Go uses Yaegi and currently only exposes `rnsd.Ret` (not `RNS`, module globals, helpers, etc.). Decide the desired interactive surface area and align UX (prompt/banner/exit commands).
- Example config parity: Go prints embedded `cmd/rnsd/example_config.txt`; Python prints `__example_rns_config__` from `rnsd.py`. Ensure these remain identical (or document intentional divergence).
- Startup/exit behaviour: Python wraps `main()` in `try/except KeyboardInterrupt` and exits cleanly; Go handles SIGINT/SIGTERM inside `programSetup()` and returns `nil` (no explicit exit code). Confirm signal/newline/exit-code behaviour matches Python expectations.
- Help/usage output: Python `argparse` formatting vs Go `flag` output will differ; decide if strict 1:1 help text is needed.

