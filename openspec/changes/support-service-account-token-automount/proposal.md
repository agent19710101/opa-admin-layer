# Proposal: support shared and topic service-account token automount control

## Why

Explicit `serviceAccountName` support closed the workload-identity binding gap, but operators still have to patch rendered Deployments when they need to disable projected Kubernetes API credentials or explicitly re-enable them for one workload. That is a common follow-up control for hardened clusters, workload identity setups, and admission-policy baselines.

## Change

- add optional shared `controlPlane.automountServiceAccountToken` to the admin spec
- add optional topic-level `automountServiceAccountToken` overrides
- inherit the shared value into rendered Deployments and let topic values replace it explicitly
- render the effective value at `Deployment.spec.template.spec.automountServiceAccountToken`
- add regression tests and example/docs coverage for decode, inheritance, override, and rendered-manifest behavior

## Impact

Operators can keep service-account identity selection and token projection policy in the same narrow renderer-owned contract without opening arbitrary pod-spec customization or relying on downstream manifest patches.
