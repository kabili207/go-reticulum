# rncp parity TODO

Only items still outstanding vs `python/RNS/Utilities/rncp.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Verify behaviour on real networks (path discovery timing, link establishment, resource strategy differences).
- Match user-facing output text 1:1 where it matters (some `flag`-generated help/usage and a few status lines will differ from Python `argparse`).
- Confirm exit codes on all failure branches match Python (`RNS.exit()` codes), not just the ones already aligned (eg. `--save` dir checks).
- Validate edge-case path handling on Windows (Python uses `os.path.abspath(os.path.expanduser(...))`; Go uses `absPath()` + `filepath`).

