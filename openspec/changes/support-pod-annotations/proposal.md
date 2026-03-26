# Proposal: support shared and topic pod annotations

## Why

The renderer now owns the Deployment pod template but still leaves no first-class path for operators to attach pod-scoped metadata such as service-mesh injection hints, tracing knobs, or admission-controller annotations. That turns a common workload integration need into downstream patching even when the rest of the deployment shape already comes from the admin spec.

The current spec already has a good inheritance model for Service metadata and OPA resources. Pod-template annotations fit the same pattern: most fleets want a shared baseline, while one tenant/topic may need a targeted override or extra annotation.

## Change

- add optional shared `controlPlane.podAnnotations` to the admin spec
- add optional topic-level `podAnnotations` overrides with the same key/value shape
- validate pod annotation keys with the existing Kubernetes metadata-key contract
- merge topic pod annotations over shared defaults during Deployment rendering
- render the effective annotations under `spec.template.metadata.annotations` in generated Deployments
- add regression tests and docs for inheritance, override, and invalid-input cases

## Impact

Operators can express narrow pod-template metadata directly in the admin spec without widening into full deployment customization. This covers common mesh and admission-controller integration points while keeping the slice small and reviewable.
