# Proposal: support shared and topic service account names

## Why

Rendered OPA Deployments still assume the default Kubernetes service account unless operators patch manifests downstream. That blocks common integrations such as projected cloud identity, narrowed RBAC, and per-workload admission policy when the control plane already owns the Deployment template.

The repository already uses a narrow inheritance model for Service metadata, resource defaults, replicas, and workload annotations/labels. Service account selection fits the same shape: most fleets want one shared baseline, while a small number of topics need targeted overrides.

## Change

- add optional shared `controlPlane.serviceAccountName` to the admin spec
- add optional topic-level `serviceAccountName` overrides
- validate both fields against the Kubernetes service-account naming contract
- inherit the shared value into rendered Deployments and let topic values override it
- render the effective value at `Deployment.spec.template.spec.serviceAccountName`
- add regression tests and example/docs coverage for decode, validation, inheritance, and override behavior

## Impact

Operators can bind rendered OPA workloads to explicit service accounts without opening arbitrary pod-spec customization or relying on downstream patches. The slice stays narrow, reviewable, and aligned with the renderer owning Deployment manifests.
