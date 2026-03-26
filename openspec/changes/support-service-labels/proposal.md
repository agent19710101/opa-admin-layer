## Why

Rendered Services already support inherited annotations, traffic policy, and session affinity, but operators still lack a narrow way to stamp Service-only labels for ownership, GitOps selection, or discovery integration. Today that forces teams to overload broader topic labels when they only want `Service.metadata.labels`.

## What Changes

- add shared `controlPlane.serviceLabels` support
- add topic-level `serviceLabels` overrides that merge over shared defaults
- validate Service label keys and values with the existing Kubernetes label contract
- render effective Service labels only onto `Service.metadata.labels`, while keeping built-in identity labels immutable
- refresh docs and example specs to show the new Service-metadata layer

## Impact

- narrows a common manifest patch point without widening into arbitrary Service customization
- keeps Deployment, ConfigMap, and pod-template metadata behavior unchanged
- preserves selector and ownership stability by refusing to let custom labels override built-in identity keys
