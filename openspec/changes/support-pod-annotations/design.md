# Design: shared and topic pod annotations

## Scope

This slice adds a narrow workload-metadata escape hatch for rendered OPA Deployments:

- shared `controlPlane.podAnnotations`
- topic-level `podAnnotations` overrides
- key validation using the existing Kubernetes metadata-key validator
- rendering only into `Deployment.spec.template.metadata.annotations`

## Rationale

Pod-template annotations are a common integration point for service meshes, trace collectors, and admission controllers. They are also lower risk than opening the full Deployment metadata surface because they do not affect selectors or object identity.

## Out of scope

- deployment-level annotations
- pod-template label overrides
- deletion semantics for inherited annotations
- arbitrary Deployment spec customization
