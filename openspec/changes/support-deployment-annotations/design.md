# Design: shared and topic deployment annotations

## Scope

This slice adds a narrow Deployment-metadata escape hatch for rendered OPA workloads:

- shared `controlPlane.deploymentAnnotations`
- topic-level `deploymentAnnotations` overrides
- key validation using the existing Kubernetes metadata-key validator
- rendering only into `Deployment.metadata.annotations`

## Rationale

Deployment-level annotations are a common integration point for rollout controllers, ownership metadata, and GitOps tooling. They are also safer than opening arbitrary Deployment spec fields because they do not affect workload identity, selector behavior, or container wiring.

## Out of scope

- arbitrary Deployment spec customization
- deployment label overrides
- deletion semantics for inherited annotations
- pod-template annotation changes beyond the existing `podAnnotations` contract
