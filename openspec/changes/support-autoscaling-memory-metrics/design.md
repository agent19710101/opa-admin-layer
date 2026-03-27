# Design: support autoscaling memory metrics

## Scope

Extend the existing autoscaling block with one additional optional field:

- `targetMemoryUtilizationPercentage`

The effective autoscaling contract becomes:

- `minReplicas`
- `maxReplicas`
- optional `targetCPUUtilizationPercentage`
- optional `targetMemoryUtilizationPercentage`

At least one utilization target must be set.

## Rationale

The current autoscaling renderer is intentionally narrow, but CPU-only metrics leave out a common Kubernetes scaling signal for OPA workloads that are driven more by loaded data/policy size than by CPU saturation. Memory utilization is the smallest useful expansion because it reuses the existing HPA v2 resource-metric shape and the existing inherited resource-request model.

Supporting CPU and memory together is also useful: operators can scale on either pressure signal without opening custom metrics or HPA behavior policy yet.

## Validation

The validation contract stays narrow:

- `minReplicas > 0`
- `maxReplicas > 0`
- `maxReplicas >= minReplicas`
- at least one of `targetCPUUtilizationPercentage` or `targetMemoryUtilizationPercentage` must be set
- each configured utilization target must be between 1 and 100
- `replicas` cannot be set on workloads governed by autoscaling
- CPU utilization requires an effective inherited `opaResources.requests.cpu`
- memory utilization requires an effective inherited `opaResources.requests.memory`

## Rendering

Rendered HPAs continue using `autoscaling/v2`. The metrics list now includes:

- a CPU resource-utilization metric when `targetCPUUtilizationPercentage` is set
- a memory resource-utilization metric when `targetMemoryUtilizationPercentage` is set

Metric order is deterministic: CPU first, then memory when both are present.

## Out of scope

- memory-value targets (`AverageValue`)
- custom or external metrics
- HPA behavior policies / stabilization windows
- partial autoscaling-field inheritance/merging
- arbitrary HPA passthrough
