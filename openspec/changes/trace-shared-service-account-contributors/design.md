# Design: trace shared ServiceAccount contributors

## Scope

This slice keeps the current shared owned `ServiceAccount` render contract and adds only plan-level traceability.

Each `sharedServiceAccounts[]` entry should include a deterministic list of contributing tenant/topic refs:

- `tenant`
- `topic`

## Rendering and API model

- collect contributor refs while grouping repeated effective `serviceAccountName` values
- sort contributor refs deterministically by tenant name, then topic name
- include the sorted refs in the in-memory plan model
- preserve the existing shared `serviceaccount.yaml` artifact shape; traceability lives in `plan.json` and REST plan responses rather than in Kubernetes manifest annotations

## Rationale

The shared-object renderer already knows which topics contributed each compatible repeated name. Surfacing those refs in the plan is the smallest useful follow-up because it improves operator auditability without taking on broader ownership semantics or mutating Kubernetes objects.

## Out of scope

- changing shared ServiceAccount compatibility validation
- adding shared object status/history metadata
- exposing contributor refs in rendered Kubernetes YAML
