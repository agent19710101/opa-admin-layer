# Proposal: support topic ConfigMap annotation overrides

## Why

Rendered ConfigMaps already support shared `controlPlane.configMapAnnotations`, but that leaves ConfigMap metadata inconsistent with the inherited override model used for Services, Deployments, pod templates, replicas, and workload identity. Operators still need per-topic ConfigMap metadata for reloader ownership, GitOps source tagging, or tenant-specific integration markers without downstream patching.

## What changes

- add optional topic-level `configMapAnnotations` to the admin spec
- validate topic annotation keys with the existing Kubernetes metadata-key contract
- merge topic `configMapAnnotations` over shared `controlPlane.configMapAnnotations` key-by-key
- render the effective annotation set into each generated ConfigMap manifest
- keep the slice intentionally narrow: ConfigMap object annotations only, with no arbitrary ConfigMap customization or deletion semantics yet

## Impact

Operators can keep shared ConfigMap metadata defaults while customizing or replacing individual annotations for one tenant/topic workload, and the ConfigMap metadata contract becomes consistent with the repository's existing inherited metadata model.
