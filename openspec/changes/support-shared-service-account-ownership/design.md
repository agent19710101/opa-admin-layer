# Design: support shared ServiceAccount ownership

## Scope

This slice adds a third effective service-account mode:

1. **no identity binding**: a topic has no effective `serviceAccountName`, so no `ServiceAccount` object is rendered
2. **unique owned mode**: exactly one topic resolves an effective `serviceAccountName`, so the renderer continues to emit one topic-owned `ServiceAccount` manifest and `serviceaccount.yaml`
3. **shared owned mode**: multiple topics resolve the same effective `serviceAccountName`, and the renderer emits one shared `ServiceAccount` object for that name when all contributing topics resolve compatible effective metadata

## Shared-object compatibility

A repeated effective name is eligible for shared owned mode only when all contributing topics resolve the same effective:

- `serviceAccountName`
- `serviceAccountAnnotations`
- `serviceAccountLabels`

If any topic disagrees on effective annotations or labels after inheritance and removals are applied, validation should fail before render with a conflict that names the shared service account and the disagreeing topics.

## Rendering model

- continue to render each topic Deployment with its effective `serviceAccountName`
- compute shared service-account ownership after topic normalization resolves inherited metadata
- emit one canonical shared `ServiceAccount` manifest per compatible repeated effective name
- write one shared `serviceaccount.yaml` artifact keyed by the shared service-account name instead of per-topic duplication
- keep renderer-owned identity labels immutable in shared mode just as they are in unique owned mode

## Rationale

The current binding-only suppression is a safe stopgap, but it still hides operator intent when the plan should own one shared identity object. Shared owned mode is the smallest useful next step because it reuses the existing service-account metadata contract, keeps Deployment wiring unchanged, and makes conflicts explicit instead of silently dropping back to manual ownership.

## Tradeoffs

This introduces cross-topic compatibility checks and a plan-level render grouping step. That is acceptable because the grouping key already exists (`serviceAccountName`) and the slice stays narrow by requiring exact effective metadata agreement rather than inventing merge precedence across topics.

## Out of scope

- merging incompatible metadata across topics
- rendering RBAC roles, role bindings, or IAM bindings
- changing the current per-topic namespace model or adding plan-wide namespace overrides
