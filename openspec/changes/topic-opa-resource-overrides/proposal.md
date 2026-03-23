# Proposal: allow per-topic OPA resource overrides

## Why

Shared `controlPlane.opaResources` defaults cover the common scheduling baseline, but they still force operators to choose between one-size-fits-all resources and downstream manifest patching. In practice, some tenant/topic workloads will need more memory or different CPU limits than the rest of the fleet.

## Change

- add optional `opaResources` to each topic in the admin spec
- validate topic-level CPU and memory quantities using the same contract as shared control-plane resources
- merge topic-level resource fields over shared `controlPlane.opaResources` defaults during plan rendering
- keep unspecified topic resource fields inherited from the shared defaults

## Impact

Operators can keep one shared OPA resource baseline while selectively tuning noisy or latency-sensitive topics in-tree. The change stays small because the renderer still emits a single Deployment resource block per topic and does not yet introduce policy budgets or auto-sizing logic.
