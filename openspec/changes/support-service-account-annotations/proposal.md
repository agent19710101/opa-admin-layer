# Proposal: support rendered ServiceAccount annotations

## Summary

Allow shared and topic-level `serviceAccountAnnotations` so rendered `ServiceAccount` manifests can carry workload-identity integration metadata without downstream patching.

## Why now

The renderer now emits `ServiceAccount` objects whenever workloads resolve an explicit effective `serviceAccountName`, but those objects still cannot carry the annotations many clusters need for IAM/workload-identity integration. That leaves a common manual patch step in the middle of the newly self-contained identity path.

## Scope

- add optional shared `controlPlane.serviceAccountAnnotations`
- add optional topic-level `serviceAccountAnnotations` overrides plus `removeServiceAccountAnnotations`
- validate annotation keys with the existing Kubernetes metadata-key contract
- render effective annotations into `ServiceAccount.metadata.annotations`

## Out of scope

- ServiceAccount labels
- imagePullSecrets, secrets, or arbitrary ServiceAccount fields
- RBAC generation or ownership/deduplication changes
