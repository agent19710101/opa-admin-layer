# Proposal: support shared and topic image pull policy control

## Why

Rendered Deployments already let operators pin the OPA image and control workload identity details, but they still have to patch manifests when one environment needs `Always`, another wants `IfNotPresent`, or a hardened/offline cluster requires `Never`. That is a narrow, common container-runtime control that fits the existing shared-plus-topic inheritance model.

## Change

- add optional shared `controlPlane.imagePullPolicy` to the admin spec
- add optional topic-level `imagePullPolicy` overrides
- validate values against the Kubernetes-supported set: `Always`, `IfNotPresent`, or `Never`
- render the effective value at `Deployment.spec.template.spec.containers[].imagePullPolicy`
- refresh examples, README, and understanding/design notes to show the new workload-runtime control

## Impact

Operators can keep image reference and image pull behavior inside the same renderer-owned Deployment contract without widening into arbitrary container customization or downstream patch workflows.
