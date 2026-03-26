# Design: inherited metadata removals

## Scope

This slice adds topic-level removal lists for inherited metadata on generated Services, ConfigMaps, Deployments, and pod templates:

- `removeServiceAnnotations`
- `removeServiceLabels`
- `removeConfigMapAnnotations`
- `removeConfigMapLabels`
- `removeDeploymentAnnotations`
- `removeDeploymentLabels`
- `removePodAnnotations`
- `removePodLabels`

## Rationale

The current inherited metadata model only supports additive merge and override. That works for many defaults, but it breaks down when a shared metadata key must be absent for one topic. Explicit removal lists are a narrow contract that solves that case without opening arbitrary manifest editing.

Applying removals after the shared/topic merge keeps behavior predictable:

1. shared metadata establishes defaults
2. topic metadata overrides shared keys when present
3. topic removal lists delete any inherited key that should end absent

Built-in identity labels remain outside that contract. They are renderer-owned and must stay immutable so selectors, ownership, and cross-object identity remain stable.

## Out of scope

- arbitrary object metadata passthrough beyond the existing object-scoped maps
- deletion semantics for broad topic labels
- more complex patch operations than key removal
- autoscaling, RBAC generation, or namespace expansion
