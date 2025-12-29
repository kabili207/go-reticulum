# go-reticulum

Go port (parity-focused) of the original Reticulum project.

This repository is maintained as a practical parity port: behaviour and public utilities are compared against the Python Reticulum reference, and remaining differences are tracked in `PARITY_*.md`. A significant part of the work was/is done with AI assistance (Codex/LLMs) for reading the reference implementation, creating parity TODOs, writing tests, and porting examples/CLI tooling.

## Goals

- Port Reticulum to Go with maximum behavioural parity.
- Keep the project testable: unit tests and selected integration tests.
- Provide CLI utilities compatible in spirit with the original ones (`rnstatus`, `rnpath`, `rnid`, etc.).

## Repository layout

- `rns/` — core library (network stack and protocols).
- `cmd/` — CLI utilities (the Go equivalents of Reticulum’s “official” tools).
- `examples/` — ports of Python examples from `python/Examples/` (small demo programs).
- `python/` — reference Python examples/materials for comparison (the “source of truth” for behaviour).

## How to run

Requires Go (recommended: current stable version).

### CLI utilities (`cmd/*`)

Quick run without building:

```bash
go run ./cmd/rnsd --help
go run ./cmd/rnstatus --help
```

Build a single tool:

```bash
go build -o ./bin/rnstatus ./cmd/rnstatus
./bin/rnstatus --help
```

### Examples (`examples/*`)

Examples are typically either “server/client” (two sides) or “standalone”.

```bash
go run ./examples/minimal -config <dir>
go run ./examples/link -server -config <dir>
go run ./examples/link -destination <hex> -config <dir>
```

### TCP uplink config (WIP)

The Go port can bring up a `TCPClientInterface` from config (enough to test uplinks).

Example snippet for `config`:

```ini
[interfaces]
  [[Uplink]]
    type = TCPClientInterface
    enabled = Yes
    target_host = reticulum.betweentheborders.com
    target_port = 4242
```

## Parity & docs

- `PARITY_*.md` files contain the remaining parity TODOs vs Python Reticulum (no unrelated wishlists).
- Ideally, each parity item is closed by a code change plus a test/verification step.
- Note: the Python Reticulum supports “external interfaces” by loading `<Type>.py` modules from `interfacepath`. The Go port does not execute/load those Python modules; if such a file exists, startup will error to avoid silently running with a placeholder interface.

## Disclaimer

This is a work-in-progress port: APIs, protocols, and CLI behaviour can change as parity improves.
