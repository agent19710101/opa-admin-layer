# Design: support HPA scaling policies and selectPolicy

## Scope

Extend the existing autoscaling behavior policy structure with two optional fields on `behavior.scaleUp` and `behavior.scaleDown`:

- `selectPolicy`
- `policies[]` with `type`, `value`, and `periodSeconds`

The existing utilization-target and stabilization-window contract stays in place.

## Rationale

After utilization targets and stabilization windows, the next smallest useful HPA behavior slice is explicit scaling-policy control. Operators commonly need to cap or prefer pod-count vs percentage-based scaling steps, and `selectPolicy` decides how multiple policies combine. That adds real production value while staying well short of full HPA passthrough.

## Validation

Validation stays intentionally narrow:

- each configured `behavior.scaleUp` / `scaleDown` block must set at least one of `stabilizationWindowSeconds`, `selectPolicy`, or `policies`
- `selectPolicy` must be one of `Max`, `Min`, or `Disabled`
- each `policies[]` entry must set `type` to `Pods` or `Percent`
- each `policies[]` entry must set `value` greater than zero
- each `policies[]` entry must set `periodSeconds` greater than zero and 1800 or fewer

Existing autoscaling validation for metric targets, replicas conflicts, and effective inherited resource requests remains unchanged.

## Rendering

Rendered HPAs continue using `autoscaling/v2`. When configured, behavior policies now also render:

- `selectPolicy`
- `policies[]` entries in spec order

Rendering remains deterministic and keeps behavior nested under the configured `scaleUp` / `scaleDown` blocks.

## Out of scope

- arbitrary HPA behavior passthrough
- custom, object, or external metrics
- partial autoscaling-field inheritance/merging
- cross-topic policy reuse or templating
