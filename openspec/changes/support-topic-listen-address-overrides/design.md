# Design: support topic listen-address overrides

## Scope

This slice keeps the existing shared `controlPlane.defaultListenAddress` contract, but lets a topic replace it with `listenAddress`.

- empty topic `listenAddress` continues to inherit the shared control-plane value
- non-empty topic `listenAddress` uses the same strict socket syntax as the shared field
- the effective topic listen address drives the rendered OPA `--addr`, container port, readiness/liveness probe port, and Service port/targetPort

## Rationale

Listen-address handling is already a narrow shared contract with good validation and renderer reuse. Extending that contract to topic scope is a small, high-leverage improvement because it removes unnecessary spec duplication while preserving the same deterministic render pipeline.

## Tradeoffs

This adds one more inherited field to topic normalization. The slice intentionally stops at single-address overrides and does not introduce broader networking customization.

## Out of scope

- multiple listeners or named ports per workload
- protocol/TLS differences between topics
- ingress, host networking, or pod port-list customization
