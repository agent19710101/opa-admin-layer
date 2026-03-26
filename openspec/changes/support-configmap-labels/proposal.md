# Change: support ConfigMap labels

## Why

Rendered ConfigMaps already support inherited annotations, but object-scoped labels still require downstream patching. That keeps ConfigMap metadata inconsistent with Deployments, pod templates, Services, and recent topic-level ConfigMap annotation support. Operators need labels for GitOps grouping, policy selection, and discovery without leaking those labels onto other rendered objects.

## What changes

- add optional shared `controlPlane.configMapLabels` to the admin spec
- add optional topic-level `configMapLabels` overrides
- validate ConfigMap label keys and values with the existing Kubernetes label contract
- merge topic `configMapLabels` over shared `controlPlane.configMapLabels`
- render the effective labels only into `ConfigMap.metadata.labels`
- keep built-in rendered identity labels immutable during ConfigMap label merging

## Impact

- operators can attach ConfigMap-only labels without post-render patches
- generated ConfigMaps follow the same inherited metadata model as other rendered objects
- Services, Deployments, and pod templates remain unchanged by ConfigMap-only label overrides
