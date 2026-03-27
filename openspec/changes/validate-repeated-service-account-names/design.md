# Design: reject repeated effective ServiceAccount names

## Scope

This slice narrows the rendered ServiceAccount contract so each effective Kubernetes ServiceAccount name has exactly one owning topic in a rendered plan.

- compute the effective `serviceAccountName` for each topic using the existing shared/topic inheritance rules
- ignore topics whose effective service-account name is empty
- fail validation when a later topic resolves the same effective name as an earlier topic
- keep plan rendering unchanged once validation passes

## Rationale

Generated ServiceAccounts already carry renderer-owned identity labels and propagated topic labels, so repeated names are not a harmless duplicate artifact problem. They represent a single Kubernetes object that would be rendered with multiple incompatible metadata payloads. Rejecting the collision is the smallest useful fix because it makes ownership explicit without introducing a cross-topic shared-object layer.

## Tradeoffs

This decision removes the current implicit ability to point multiple topics at the same rendered ServiceAccount name. Operators that truly need shared workload identity must wait for a future shared-object ownership model or keep ServiceAccount provisioning outside this renderer.

## Out of scope

- deduplicated shared ServiceAccount objects in plan/export output
- deterministic merge rules for repeated ServiceAccount metadata
- Deployment-level deduplication or cross-topic workload grouping beyond ServiceAccount ownership
