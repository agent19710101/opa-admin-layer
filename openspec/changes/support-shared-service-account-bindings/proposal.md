# Proposal: support shared service-account bindings

## Summary

Allow repeated effective `serviceAccountName` values across topics as binding-only references while suppressing rendered `ServiceAccount` manifests for those shared names.

## Why now

The previous contract forced operators to choose between renderer-managed `ServiceAccount` creation and intentionally reusing one existing/shared Kubernetes service account across multiple workloads. In practice, reusing one service account is a common workload-identity pattern, and rejecting it outright makes the admin layer less useful than plain handwritten manifests.

## Scope

- allow repeated effective `serviceAccountName` values to pass validation
- keep `Deployment.spec.template.spec.serviceAccountName` rendering unchanged for those workloads
- omit rendered `ServiceAccount` YAML/artifacts for effective names used by more than one topic
- document the binding-only shared-name behavior in README, architecture notes, understanding docs, and changelog

## Out of scope

- deduplicated shared `ServiceAccount` manifests owned once per plan
- merge/conflict rules for shared service-account annotations or labels across topics
- RBAC generation or broader workload identity customization
