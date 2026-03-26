# Design: support OPA autoscaling

## Scope

This slice introduces an optional autoscaling contract for generated OPA Deployments:

- shared `controlPlane.autoscaling`
- topic-level `autoscaling` overrides
- rendered `HorizontalPodAutoscaler` manifests targeting the generated Deployment

The autoscaling block is intentionally narrow:

- `minReplicas`
- `maxReplicas`
- `targetCPUUtilizationPercentage`

## Rationale

The current workload-control contract stops at fixed replicas. That is enough for simple environments, but it still pushes a common Kubernetes scaling behavior into downstream patches. A small autoscaling block lets the renderer own the HPA manifest alongside the Deployment it already owns.

Topic-level autoscaling replaces the shared autoscaling block instead of merging field-by-field. That keeps override behavior obvious and avoids partial effective HPA shapes.

When autoscaling is configured, the rendered Deployment replica count follows `minReplicas`. That gives the workload a deterministic starting size while letting the HPA controller own future scaling behavior.

## Validation

Autoscaling requires all three fields and validates them narrowly:

- `minReplicas > 0`
- `maxReplicas > 0`
- `maxReplicas >= minReplicas`
- `targetCPUUtilizationPercentage` between 1 and 100
- `replicas` cannot be set on workloads governed by autoscaling

## Out of scope

- memory or custom metrics
- behavior policies, stabilization windows, or scale rules
- partial autoscaling-field inheritance/merging
- arbitrary Deployment or HPA passthrough
- KEDA or other external autoscaling integrations