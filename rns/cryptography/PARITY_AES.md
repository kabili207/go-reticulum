# aes parity TODO

Only items still outstanding vs `python/RNS/Cryptography/AES.py` (everything else is already ported and/or covered by unit tests).

## TODO (remaining parity gaps)

- Confirm exact error strings/exception mapping (Python raises exceptions; Go returns `error` with slightly different wording).
- Verify behaviour on non-multiple-of-block inputs matches Python call sites (Go requires callers to PKCS7-pad first).

