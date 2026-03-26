# Proposal: support shared and topic pod labels

## Why

The renderer already propagates topic labels into workload metadata, but operators still lack a first-class path to add pod-template-specific labels for mesh, admission, scheduling, or observability integrations without also stamping those labels onto ConfigMaps, Services, and Deployment metadata. That keeps a common pod-scoped integration need stuck in downstream patching.

The spec already uses a narrow inheritance model for Service metadata, resource defaults, and annotations. Pod-template labels fit the same model: most fleets want a shared baseline, while one tenant/topic may need a targeted override.

## Change

- add optional shared `controlPlane.podLabels` to the admin spec
- add optional topic-level `podLabels` overrides with the same key/value shape
- validate pod label keys and values with the existing Kubernetes label contract
- merge topic pod labels over shared defaults during Deployment rendering
- render the effective labels under `spec.template.metadata.labels` in generated Deployments
- keep built-in identity labels immutable so pod-template overrides cannot break Deployment selectors
- add regression tests and docs for inheritance, override, and invalid-input cases

## Impact

Operators can express narrow pod-template label metadata directly in the admin spec without widening into full Deployment customization or leaking pod-only labels onto every rendered object. This covers common mesh and workload-integration needs while preserving selector safety.
