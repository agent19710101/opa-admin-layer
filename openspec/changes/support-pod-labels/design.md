# Design: shared and topic pod labels

## Scope

This slice adds a narrow pod-template metadata escape hatch for rendered OPA workloads:

- shared `controlPlane.podLabels`
- topic-level `podLabels` overrides
- key/value validation using the existing Kubernetes label validators
- rendering only into `Deployment.spec.template.metadata.labels`
- selector-safe merging that preserves the built-in identity labels already used by the Deployment selector and Service selector

## Rationale

Pod-template labels are a common integration point for mesh sidecar selection, workload policy attachment, monitoring discovery, and scheduling metadata. They are also safer than opening arbitrary pod spec customization because the surface stays limited to labels while the renderer still owns containers, probes, volumes, and selectors.

The renderer already carries built-in identity labels and topic labels across multiple manifests. This slice intentionally keeps pod-scoped labels separate so operators can add pod-only metadata without automatically stamping the same labels onto Services or ConfigMaps.

## Out of scope

- deployment-level label overrides
- deletion semantics for inherited pod labels
- arbitrary pod spec customization
- selector changes or built-in identity label overrides
