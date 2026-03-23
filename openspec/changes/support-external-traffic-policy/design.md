# Design: Service externalTrafficPolicy support

## Context

Generated Services already support a shared `serviceType`, shared `serviceAnnotations`, and topic-level overrides for both. Operators still need one more narrow networking control in common deployments: `externalTrafficPolicy`, especially when a Service is exposed via `NodePort` or `LoadBalancer`.

## Decision

Add `externalTrafficPolicy` to both the shared control-plane block and each topic with the same inheritance pattern used by other Service metadata:

1. Start from `controlPlane.externalTrafficPolicy`.
2. If `topic.externalTrafficPolicy` is set, use it instead for that topic.
3. Accept only `Cluster` and `Local`.
4. Revalidate the effective `serviceType` + `externalTrafficPolicy` pair so traffic policy is emitted only when the final Service type is `NodePort` or `LoadBalancer`.

## Consequences

Positive:

- removes another common Service patch step from operator workflows
- keeps Service metadata controls consistent across shared defaults and topic overrides
- adds a useful load-balancer/node-local routing knob without expanding into broader networking abstractions

Tradeoffs:

- validation now depends on the effective merged Service shape, not just raw field syntax
- deeper Service/network controls (extra ports, internal traffic policy, ingress/mesh features) remain intentionally out of scope
