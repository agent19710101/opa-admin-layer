# Proposal: support rendered ServiceAccount labels

## Summary

Allow shared and topic-level `serviceAccountLabels` so rendered `ServiceAccount` manifests can carry ownership, GitOps, and policy labels without downstream patching.

## Why now

The renderer now emits `ServiceAccount` objects and supports inherited ServiceAccount annotations, but those generated identities still cannot carry labels. That leaves a small metadata gap for clusters that rely on labels for ownership, policy, or inventory workflows.

## Scope

- add optional shared `controlPlane.serviceAccountLabels`
- add optional topic-level `serviceAccountLabels` overrides plus `removeServiceAccountLabels`
- validate label keys and values with the existing Kubernetes label contract
- render effective labels into `ServiceAccount.metadata.labels`

## Out of scope

- RBAC generation or ownership/deduplication changes
- imagePullSecrets, secrets, or arbitrary ServiceAccount fields
- arbitrary ServiceAccount passthrough
