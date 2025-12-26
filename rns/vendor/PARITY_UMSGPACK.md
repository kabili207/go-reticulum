# umsgpack parity TODO

Only items still outstanding vs `python/RNS/vendor/umsgpack.py` (u-msgpack-python v2.7.1).

## TODO (remaining parity gaps)

- Type coverage: Go implementation intentionally supports a subset (nil/bool/ints/floats/bytes/strings/arrays/maps) and does not implement `Ext`, timestamps, and other advanced types from upstream Python module.
- Numeric fidelity: Go decodes most integers into `int64` (or `uint64` for very large unsigned) whereas Python preserves `int` semantics; ensure higher layers tolerate this.
- Map key handling: Go converts decoded `[]byte` keys to `string` to keep keys comparable; confirm this matches all Reticulum payload expectations.

