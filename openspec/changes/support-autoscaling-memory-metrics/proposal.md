# Proposal: support autoscaling memory metrics

## Why

The current HPA contract is CPU-only. That works for many workloads, but OPA can also be memory-bound under larger policy/data sets. Without a memory metric in the spec, operators still have to patch generated HPAs downstream for a common Kubernetes autoscaling case.

## Change

- extend shared and topic-level `autoscaling` blocks with optional `targetMemoryUtilizationPercentage`
- allow autoscaling to target CPU, memory, or both resource-utilization metrics
- require at least one autoscaling utilization target to be set
- validate effective inherited `opaResources.requests.memory` when memory utilization is configured
- render a memory resource metric into generated `HorizontalPodAutoscaler` manifests when requested

## Impact

Operators can keep using the current narrow inherited autoscaling contract while covering memory-sensitive OPA workloads. The slice stays small and reviewable because it still avoids behavior policies, custom metrics, and arbitrary HPA passthrough.
