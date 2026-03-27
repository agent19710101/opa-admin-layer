# Proposal: trace shared ServiceAccount contributors

## Why

Shared owned `ServiceAccount` rendering now deduplicates compatible repeated effective `serviceAccountName` values into one plan-scoped object, but the resulting `plan.json` no longer shows which tenant/topic workloads contributed that shared identity. Operators need that traceability to review ownership before apply time.

## What changes

- add contributing tenant/topic references to each `sharedServiceAccounts[]` plan entry
- expose the same contributor list through CLI `render` output (`plan.json`) and REST `/v1/plans`
- document the new traceability contract for shared owned ServiceAccounts

## Out of scope

- RBAC or lifecycle ownership generation
- broader GitOps status/ownership metadata beyond contributing topic refs
- changes to shared ServiceAccount compatibility rules
