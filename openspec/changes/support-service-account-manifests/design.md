# Design: rendered ServiceAccount manifests

## Scope

This slice extends the existing workload-identity contract just enough to keep generated deployment bundles self-contained when operators opt into explicit service-account binding.

- reuse the existing shared/topic `serviceAccountName` inheritance contract
- when the effective service-account name is non-empty, emit a Kubernetes `ServiceAccount` manifest for that topic
- expose the rendered YAML in the topic plan and write `serviceaccount.yaml` during `render -outdir`

## Rationale

The current renderer already owns adjacent workload objects such as ConfigMaps, Deployments, Services, and optional HPAs. Rendering the matching ServiceAccount is the smallest useful follow-up because it removes a common manual prerequisite without widening into RBAC policy generation.

## Tradeoffs

The first slice renders ServiceAccounts per topic, even if multiple topics converge on the same effective name. That keeps the plan model consistent with existing per-topic artifacts and avoids introducing a new cross-topic shared-object layer during this slice.

## Out of scope

- Role/ClusterRole/RoleBinding generation
- ServiceAccount metadata customization
- deduplicating or centrally listing shared ServiceAccounts across topics
