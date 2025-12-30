`configs/` contains ready-to-use configs and a "kitchen sink" template for Reticulum (Go port).

The format matches Python: it is a directory that contains a `config` file.
All utilities accept `-config <dir>` (ie. a path to the directory, not the file).

## Quick start

- Start a single instance: `go run ./cmd/rnsd -config ./configs/testing/single_shared_tcp`
- View status: `go run ./cmd/rnstatus -a -config ./configs/testing/single_shared_tcp`

## Structure

- `configs/kitchen_sink/` — "all-in-one" reference (all relevant keys/interfaces, mostly `enabled = no`).
- `configs/testing/` — small deterministic configs convenient for `tests/integration/*.sh`.

## Multiple instances

To run multiple `rnsd` instances at the same time:
- use different `instance_name`
- and (preferably) `shared_instance_type = tcp` + unique `shared_instance_port` / `instance_control_port`
- also separate interface ports (UDP/TCP) to avoid collisions.

Ready-made pairs:
- `configs/testing/two_nodes_udp/node_a` + `configs/testing/two_nodes_udp/node_b`
- `configs/testing/two_nodes_tcp/server` + `configs/testing/two_nodes_tcp/client`
