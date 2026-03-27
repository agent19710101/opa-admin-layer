# Proposal: support rendered ServiceAccount manifests

## Summary

Close the current workload-identity gap between binding and provisioning by rendering a Kubernetes ServiceAccount manifest whenever a workload resolves an explicit `serviceAccountName`.

## Why now

The repository already lets operators bind generated Deployments to shared or topic-specific service account names, but the plan still leaves those identities to be created manually. That creates an avoidable integration gap for the common case where the admin layer is already expected to emit the workload-scoped Kubernetes objects it owns.

## Scope

- render a `ServiceAccount` manifest per topic when the effective inherited `serviceAccountName` is non-empty
- include that manifest in plan JSON and `render -outdir` output
- keep the slice intentionally narrow: no RBAC, no annotations/labels, no ownership/deduplication logic beyond per-topic render output

## Out of scope

- Role or RoleBinding generation
- ServiceAccount annotations, labels, image pull secrets, or arbitrary pod-spec customization
- cross-topic deduplication or shared-object ownership management
