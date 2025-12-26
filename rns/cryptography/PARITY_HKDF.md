# hkdf parity TODO

Only items still outstanding vs `python/RNS/Cryptography/HKDF.py` (everything else is already ported and/or covered by unit tests).

## TODO (remaining parity gaps)

- Ensure counter wrap/length limits match Python exactly (Python uses `(i+1)%(0xFF+1)`; Go `HKDF()` matches, `HKDFSHA256/512` enforces `<=255` blocks).
- Confirm defaulting behaviour for `salt` and `context/info` matches Python on all call sites.

