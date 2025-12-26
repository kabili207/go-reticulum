# ed25519 parity TODO

Only items still outstanding vs `python/RNS/Cryptography/Ed25519.py` (everything else is already ported and/or covered by unit tests).

## TODO (remaining parity gaps)

- Confirm key material format expectations match Python (seed vs private key bytes) across all call sites.
- Align error types/messages for invalid key lengths and invalid signatures (Python raises; Go returns `error`).

