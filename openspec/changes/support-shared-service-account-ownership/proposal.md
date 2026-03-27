# Proposal: support shared ServiceAccount ownership

## Summary

Allow the plan renderer to emit one shared `ServiceAccount` manifest per repeated effective `serviceAccountName` when all contributing topics agree on the shared object metadata.

## Why now

The current shared-binding contract solved the duplicate-ownership bug by suppressing `ServiceAccount` output for repeated names, but that also means the plan cannot own the common case where several topics intentionally share one renderer-managed identity. Operators now have to choose between duplicated per-topic names or external/manual service-account management even when the metadata is compatible.

## Scope

- introduce a plan-scoped shared-object mode for repeated effective `serviceAccountName` values
- emit at most one rendered `ServiceAccount` manifest and `serviceaccount.yaml` artifact per shared effective name
- require shared topics to agree on effective shared-object metadata before render
- keep `Deployment.spec.template.spec.serviceAccountName` rendering unchanged for all topics
- document the shared-ownership behavior in README, architecture notes, understanding docs, and changelog

## Out of scope

- RBAC generation or policy attachment for shared service accounts
- cross-namespace shared service-account ownership
- arbitrary shared-object customization beyond the existing service-account metadata contract
