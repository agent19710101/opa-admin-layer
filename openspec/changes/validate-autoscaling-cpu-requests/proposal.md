# Proposal: validate autoscaling CPU requests

## Why

The new autoscaling slice emits Kubernetes HorizontalPodAutoscaler manifests using CPU utilization targets. That contract is incomplete when the effective OPA container resources omit `requests.cpu`, because Kubernetes cannot calculate CPU utilization targets reliably without a CPU request baseline.

## Change

- validate that any workload with effective `autoscaling` also has an effective `opaResources.requests.cpu` value after shared/topic inheritance is applied
- keep the validation narrow and tied only to CPU-targeted autoscaling
- document the requirement in README and architecture/understanding notes

## Impact

Operators get an earlier, clearer validation failure instead of rendering an HPA shape that is unlikely to work correctly at apply/runtime. The inherited validation model also stays consistent: shared defaults can satisfy the requirement, or a topic can provide its own CPU request when it opts into autoscaling.
