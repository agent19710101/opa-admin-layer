# Design: support shared service-account bindings

## Scope

This slice separates two service-account modes:

1. **owned render mode**: one topic resolves a unique effective `serviceAccountName`, so the renderer continues to emit a `ServiceAccount` manifest and `serviceaccount.yaml`
2. **shared binding mode**: multiple topics resolve the same effective `serviceAccountName`, so each rendered Deployment still references that name, but the renderer omits `ServiceAccount` manifest output for all of those topics

## Rationale

The previous validation error protected the renderer from claiming one Kubernetes object with multiple topic identities, but it also blocked a common operator pattern: several workloads intentionally binding to one already-managed service account. Treating repeated names as binding-only is the smallest useful improvement because it restores that deployment pattern without inventing a cross-topic shared-object ownership model.

## Tradeoffs

Operators lose renderer-managed `ServiceAccount` creation for repeated names. That is intentional: the plan should not claim ownership of a shared object until the repository has a real shared-object model with explicit metadata merge rules.

## Out of scope

- one canonical shared `ServiceAccount` object emitted once per plan
- shared-name metadata compatibility checks or merge semantics
- RBAC roles, bindings, or token-secret generation
