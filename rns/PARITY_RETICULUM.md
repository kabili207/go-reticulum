# PARITY: Reticulum (python/RNS/Reticulum.py ↔ rns/reticulum.go)

## TODO

- External interfaces: Python can load `<Type>.py` modules from `interfacepath` at runtime. The Go port does not execute/load Python modules. If a matching `<Type>.py` exists under `InterfacePath`, startup errors explicitly to avoid silently running with a placeholder interface.
- Daemon management API: `HaltInterface/ResumeInterface/ReloadInterface` are currently no-ops.
