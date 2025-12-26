# x25519 parity TODO

Only items still outstanding vs `python/RNS/Cryptography/X25519.py` and timing-equalisation behaviour.

## TODO (remaining parity gaps)

- Timing equalisation: Go implements a wall-clock delay window similar to Python; verify behaviour under load and across platforms (avoid flaky tests).
- Confirm clamping semantics and public key byte order match Python expectations.

