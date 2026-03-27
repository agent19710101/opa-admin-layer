# Proposal: support HPA stabilization windows

## Why

The autoscaling contract now covers CPU and memory utilization targets, but it still cannot express one of the most common operator-side tuning needs: slowing scale-up or scale-down reactions to avoid thrash. Without a narrow behavior field, generated HPAs still need downstream patching for a common production control.

## Change

- extend shared and topic-level `autoscaling` blocks with an optional `behavior` section
- allow `behavior.scaleUp.stabilizationWindowSeconds`
- allow `behavior.scaleDown.stabilizationWindowSeconds`
- validate empty behavior objects and invalid stabilization-window values early
- render the configured behavior block into generated `HorizontalPodAutoscaler` manifests

## Impact

Operators gain a useful HPA tuning knob without opening arbitrary behavior-policy passthrough, custom metrics, or broader scaling-rule configuration.
