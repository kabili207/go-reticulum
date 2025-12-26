# platformutils parity TODO

Only items still outstanding vs `python/RNS/vendor/platformutils.py`.

## TODO (remaining parity gaps)

- `use_af_unix()`: Python returns true only for linux/android; Go also enables it on darwin. Decide whether this is an intentional Go extension or should match Python strictly.
- `platform_checks()`: Python validates minimum Python version on Windows and calls `RNS.panic()`; Go currently no-ops on Windows. Decide whether to add an equivalent guard (or document as not applicable).
- `cryptography_old_api()`: Python inspects `cryptography.__version__`; Go always returns false. Decide whether callers need an equivalent capability flag.

