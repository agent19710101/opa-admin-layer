# Proposal: render OPA deployment health probes

## Why

The generated Deployment is now self-contained for configuration, but it still omits container port metadata and Kubernetes health probes. That weakens rollout feedback, makes restarts less deterministic, and leaves operators to patch basic runtime safety into generated manifests after render.

## Change

- derive the OPA HTTP port from the rendered listen address
- declare that port in the generated container spec
- add a default readiness probe that checks the OPA health endpoint with plugin readiness semantics
- add a default liveness probe that checks the base OPA health endpoint
- add regression coverage for default and explicit listen-address cases

## Impact

Rendered manifests become more runnable in real clusters without adding new required spec fields. Operators get better rollout visibility and safer default restart behavior while keeping the current OPA-only deployment shape.