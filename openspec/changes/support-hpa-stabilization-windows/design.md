# Design: support HPA stabilization windows

## Scope

Extend the existing autoscaling block with one optional nested behavior structure:

- `behavior.scaleUp.stabilizationWindowSeconds`
- `behavior.scaleDown.stabilizationWindowSeconds`

The existing autoscaling metric contract stays unchanged.

## Rationale

Stabilization windows are the smallest useful next HPA behavior slice after CPU/memory utilization metrics. They cover a real operator need — reducing scale oscillation — while keeping the contract much narrower than full HPA policy passthrough.

## Validation

The validation contract stays intentionally small:

- `behavior` must configure `scaleUp` and/or `scaleDown`
- each configured policy must set `stabilizationWindowSeconds`
- `stabilizationWindowSeconds` must be zero or greater
- `stabilizationWindowSeconds` must be 3600 or fewer

The existing autoscaling validation remains in force for replicas, metric targets, and effective inherited resource requests.

## Rendering

Rendered HPAs continue using `autoscaling/v2` and keep the deterministic metric order.
When behavior is configured, rendering appends:

- `behavior.scaleUp.stabilizationWindowSeconds` when configured
- `behavior.scaleDown.stabilizationWindowSeconds` when configured

## Out of scope

- HPA `policies`
- HPA `selectPolicy`
- arbitrary autoscaling behavior passthrough
- partial autoscaling-field inheritance/merging
- custom or external metrics
