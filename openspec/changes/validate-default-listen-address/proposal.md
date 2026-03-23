# Proposal: validate control-plane default listen address

## Why

The admin contract currently treats `controlPlane.defaultListenAddress` as optional, but once an operator provides a non-empty value the renderer still accepts malformed addresses and silently falls back to port `8181` when deriving Deployment probe and Service ports. That creates a drift risk where the generated `--addr` argument, container port, readiness/liveness probes, and Service target can disagree about which socket OPA is actually expected to bind.

## Change

- validate `controlPlane.defaultListenAddress` when provided instead of only normalizing the empty default
- accept the same listen-address shapes the renderer already intends to support for OPA (`:port`, `host:port`, and bracketed IPv6 host:port)
- reject malformed, portless, non-numeric, and out-of-range listen addresses through the shared validation path used by CLI and REST flows
- add regression coverage for valid and invalid listen-address shapes

## Impact

Probe and Service port derivation stay aligned with the configured OPA listen socket, and operator mistakes fail during validation instead of producing superficially valid manifests that can route traffic to the wrong port.
