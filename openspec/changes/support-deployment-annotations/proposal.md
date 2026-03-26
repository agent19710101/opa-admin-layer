# Proposal: support shared and topic deployment annotations

## Why

The renderer already owns full Kubernetes Deployment manifests, but operators still have no first-class path to attach metadata to the Deployment object itself. Pod-template annotations cover mesh and admission hooks, yet rollout tooling, ownership markers, and GitOps coordination commonly rely on top-level Deployment annotations.

Leaving that gap forces downstream patching even when the rest of the workload shape already comes from the admin spec.

## Change

- add optional shared `controlPlane.deploymentAnnotations` to the admin spec
- add optional topic-level `deploymentAnnotations` overrides with the same key/value shape
- validate deployment annotation keys with the existing Kubernetes metadata-key contract
- merge topic deployment annotations over shared defaults during Deployment rendering
- render the effective annotations under `Deployment.metadata.annotations`
- add regression tests and docs for inheritance, override, and invalid-input cases

## Impact

Operators can express narrow Deployment-object metadata directly in the admin spec without widening into arbitrary Deployment customization. This covers common rollout, ownership, and GitOps integration points while staying aligned with the existing inheritance model used by Service metadata, OPA resources, and pod-template annotations.
