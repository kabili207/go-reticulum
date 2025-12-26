# hmac parity TODO

Only items still outstanding vs `python/RNS/Cryptography/HMAC.py`.

## TODO (remaining parity gaps)

- Confirm blocksize fallback logic matches Python for non-standard hashes (Go enforces `>=16` then falls back to 64 like Python).
- Validate any subtle differences around incremental updates vs Python behaviour (we store concatenated `data` and compute on demand).

