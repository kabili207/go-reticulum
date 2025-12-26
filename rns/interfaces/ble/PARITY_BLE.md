# BLETransport parity TODO

Only items still outstanding vs Python BLE transport implementation (Reticulum’s `BLEConnection` in `python/RNS/Interfaces/RNodeInterface.py`) (everything else is already ported and/or covered by unit/integration tests).

## TODO (remaining parity gaps)

- Android parity: Python has native Android BLE; Go still has no Android BLE backend (stub). Implementing this requires an Android-specific backend (gomobile/bindings to Android BLE APIs).

