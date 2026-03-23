# Proposal: enforce OPA resource request/limit budgets

## Why

The admin layer now validates Kubernetes quantity syntax for shared and topic-level `opaResources`, but it still allows request/limit combinations that Kubernetes will reject or normalize unexpectedly. The most important guardrail is still missing: a request must not exceed the corresponding limit, even after topic overrides merge over shared defaults.

## Change

- validate shared `controlPlane.opaResources` so CPU and memory requests cannot exceed their matching limits
- validate the effective topic-level OPA resource profile after shared defaults and topic overrides are merged
- return field-specific validation errors that explain the offending request/limit pair
- add regression coverage for admin, CLI, and REST validation paths
- update README, architecture notes, and understanding docs to describe the stricter inherited resource-budget contract

## Impact

Operators get earlier feedback for invalid OPA resource budgets, and per-topic overrides can no longer silently render Deployments with request/limit combinations that the cluster would reject at apply time.
