# Identity parity TODO

Only items still outstanding vs `python/RNS/Identity.py` (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- `known_destinations` persistence format: Python uses msgpack with bytes keys; Go persists with string keys (raw bytes) but can read both.
