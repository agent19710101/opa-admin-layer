## Why

Rendered Deployments already support inherited annotations and pod-template labels, but operators still lack a narrow way to stamp deployment-only labels for rollout tracking, ownership, or GitOps selection. Today that forces teams to overload broader topic labels or pod labels when they only want `Deployment.metadata.labels`.

## What Changes

- add shared `controlPlane.deploymentLabels` support
- add topic-level `deploymentLabels` overrides that merge over shared defaults
- validate deployment label keys and values with the existing Kubernetes label contract
- render effective deployment labels only onto `Deployment.metadata.labels`, while keeping built-in identity labels immutable
- refresh docs and example specs to show the new workload-metadata layer

## Impact

- narrows a common manifest patch point without widening into arbitrary Deployment customization
- keeps Service, ConfigMap, and pod-template metadata behavior unchanged
- preserves selector and ownership stability by refusing to let custom labels override built-in identity keys
