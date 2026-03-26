# Proposal: support OPA autoscaling

## Why

The renderer now covers fixed shared/topic `replicas`, but that still leaves a common production workload control gap: operators cannot express a Kubernetes HorizontalPodAutoscaler for generated OPA Deployments without downstream patching.

## Change

- add an optional shared `controlPlane.autoscaling` block with `minReplicas`, `maxReplicas`, and `targetCPUUtilizationPercentage`
- allow topic-level `autoscaling` overrides using the same narrow shape
- render a Kubernetes `HorizontalPodAutoscaler` manifest when autoscaling is configured
- keep the slice intentionally narrow by supporting CPU utilization targets only and by making `replicas` incompatible with autoscaling-managed workloads

## Impact

Operators can move a generated workload from fixed replica counts to a first-class HPA contract without reopening arbitrary Deployment customization. The scope stays small, reviewable, and consistent with the existing inherited configuration model.