# Proposal: configure rendered Service type

## Why

The renderer now emits a Kubernetes Service for every tenant/topic OPA workload, but it always hard-codes `ClusterIP`. That is fine for in-cluster consumers, yet it still forces operators to patch generated YAML when the same plan needs node-level or load-balancer exposure.

## Change

- add optional `controlPlane.serviceType` to the admin spec
- default the rendered Service type to `ClusterIP` when the field is omitted
- support a small safe allowlist: `ClusterIP`, `NodePort`, and `LoadBalancer`
- reject unsupported values during validation so CLI and REST flows fail before render
- add regression tests and docs for the new control-plane knob

## Impact

Generated plans stay minimal by default while covering the most common Kubernetes service exposure modes without downstream manifest patches.
