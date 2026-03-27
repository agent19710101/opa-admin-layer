# Proposal: support HPA scaling policies and selectPolicy

## Why

The autoscaling contract now covers utilization targets and stabilization windows, but it still cannot express the other common HPA behavior knob: how aggressively scale-up or scale-down should proceed. Without narrow `selectPolicy` and `policies` support, generated HPAs still need downstream patching for basic step-size and selection behavior.

## Change

- extend shared and topic-level `autoscaling.behavior.scaleUp` / `scaleDown` blocks with optional `selectPolicy`
- extend those behavior blocks with optional `policies` entries (`type`, `value`, `periodSeconds`)
- validate unsupported select-policy values, empty policy entries, and invalid scaling-policy fields early
- render configured behavior policy selection and scaling policies into generated `HorizontalPodAutoscaler` manifests

## Impact

Operators gain a practical HPA tuning slice for step-size and selection behavior without opening full autoscaling passthrough, custom metrics, or arbitrary controller-specific fields.
