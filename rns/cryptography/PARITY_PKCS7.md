# pkcs7 parity TODO

Only items still outstanding vs `python/RNS/Cryptography/PKCS7.py` (everything else is already ported and/or covered by unit tests).

## TODO (remaining parity gaps)

- Confirm edge-case error messages/exception mapping (Python raises; Go returns `error`).
- Validate block size defaulting behaviour remains aligned (Go treats `bs<=0` as default).

