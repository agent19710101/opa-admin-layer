# Proposal: support topic-specific Service overrides

## Why

The renderer now owns a full Kubernetes Service manifest and already supports shared `controlPlane.serviceType` and `controlPlane.serviceAnnotations`. That covers the common baseline, but it still forces downstream patching when one tenant/topic needs a different exposure mode or controller metadata than the rest of the fleet.

In practice, operators often need one topic to stay internal while another uses a public load balancer, or to attach topic-specific annotations for cloud load balancer classes, health checks, or mesh/controller integration. Keeping Service metadata shared-only turns topic boundaries into a deployment-time patch problem instead of a first-class admin contract.

## Change

- add optional `serviceType` and `serviceAnnotations` fields to each topic in the admin spec
- validate topic-level `serviceType` with the same `ClusterIP`/`NodePort`/`LoadBalancer` contract already used for shared defaults
- validate topic-level annotation keys with the same Kubernetes metadata-key checks already used for shared Service annotations
- merge topic-level annotations over shared `controlPlane.serviceAnnotations` defaults during plan rendering
- treat topic-level `serviceType` as a full override of the shared control-plane Service type for that tenant/topic
- add regression tests and docs for inheritance, override, merge, and invalid-input cases

## Impact

Operators keep a compact shared networking baseline while gaining a targeted escape hatch for topic-specific Service exposure and metadata without cloning specs or patching rendered YAML.