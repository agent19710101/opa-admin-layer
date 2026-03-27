# Design: validate autoscaling CPU requests

## Scope

This slice adds one guardrail to the existing autoscaling contract:

- if a workload has effective autoscaling, its effective `opaResources.requests.cpu` must be set

## Rationale

The renderer currently supports only CPU utilization HPAs. CPU utilization is measured relative to requested CPU, so allowing autoscaling without a CPU request produces a weak operator contract even if the HPA manifest itself can still be rendered.

The validation should run on the effective inherited workload shape, not only raw shared/topic input:

- shared `controlPlane.opaResources.requests.cpu` should satisfy the requirement for all autoscaled topics
- a topic without shared CPU requests can still satisfy the requirement by setting its own `opaResources.requests.cpu`
- topic overrides remain replace/merge based; no new autoscaling merge behavior is introduced

## Out of scope

- memory or custom-metric autoscaling
- new HPA behavior/policy fields
- automatic defaulting of CPU requests
- rejecting non-autoscaled workloads that omit CPU requests
