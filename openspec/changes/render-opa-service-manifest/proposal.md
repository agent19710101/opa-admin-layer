# Proposal: render OPA Service manifests

## Why

The generated plan now provisions OPA configuration and a runnable Deployment, but operators still need to hand-author a Kubernetes Service to reach the per-topic OPA API on a stable in-cluster DNS name. That keeps the first slice slightly incomplete and pushes routine manifest glue work downstream.

## Change

- render a Kubernetes Service manifest for each tenant/topic OPA deployment
- use the derived OPA listen port for the Service port and targetPort
- export the Service manifest in the plan tree beside the other per-topic artifacts
- add regression tests and docs for the new artifact

## Impact

Rendered plans become more directly usable in Kubernetes by including the standard network entrypoint for each generated OPA workload without introducing new required spec fields.
