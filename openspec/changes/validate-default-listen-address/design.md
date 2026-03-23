# Design: control-plane default listen address validation

## Problem

The renderer derives the OPA container port, probe ports, and Service ports from `controlPlane.defaultListenAddress`, but the validation layer does not currently verify that a user-supplied non-empty value is parseable. The port-derivation helper therefore falls back to `8181` on malformed input while the rendered OPA args still preserve the malformed `--addr` value.

Example failure mode:

- input: `controlPlane.defaultListenAddress: localhost`
- rendered OPA arg: `--addr=localhost`
- rendered container/probe/Service ports: `8181`

That plan is internally inconsistent and will fail later than necessary.

## Decision

Treat non-empty `controlPlane.defaultListenAddress` values as part of the validation contract.

Accepted shapes for the first validation slice:

- `:8181`
- `127.0.0.1:8282`
- `[::1]:8181`

Rejected shapes:

- `localhost` (missing port)
- `8181` (missing separator/host:port form)
- `:abc` (non-numeric port)
- `127.0.0.1:70000` (out-of-range port)

Empty values remain allowed and continue to normalize to `:8181`.

## Consequences

- CLI and REST validation will reject malformed listen addresses before plan rendering.
- Probe, Service, and container port derivation now reuse the same strict parser as validation and default normalization instead of keeping a separate best-effort fallback path in the renderer.
- Future listen-address expansions such as Unix sockets or scheme-bearing URLs will require an intentional contract update instead of being partially accepted by accident.
