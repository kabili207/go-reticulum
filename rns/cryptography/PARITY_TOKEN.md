# token parity TODO

Only items still outstanding vs `python/RNS/Cryptography/Token.py` (everything else is already ported and/or covered by unit tests).

## TODO (remaining parity gaps)

- Confirm exact token binary format matches Python: `IV || ciphertext || HMAC-SHA256` and key splitting semantics.
- Validate error handling/exception mapping for invalid tokens and invalid padding (Go wraps errors; Python raises).

