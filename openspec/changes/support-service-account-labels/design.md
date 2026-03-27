# Design: rendered ServiceAccount labels

## Scope

This slice extends the existing inherited metadata model to ServiceAccount object labels.

- reuse the existing shared/topic inheritance pattern already used for Service, ConfigMap, Deployment, and pod labels
- allow topics to override shared keys and explicitly remove inherited keys through `removeServiceAccountLabels`
- preserve the renderer-owned built-in identity labels while still carrying propagated topic labels onto generated ServiceAccounts
- render labels only into `ServiceAccount.metadata.labels`

## Rationale

ServiceAccount labels are the next smallest useful workload-identity metadata step after ServiceAccount annotations because many GitOps, ownership, and policy workflows key off labels rather than annotations. Keeping the slice to labels only closes that gap without widening into RBAC generation or arbitrary ServiceAccount customization.

## Tradeoffs

Generated ServiceAccounts now carry the same built-in identity labels and topic labels as other rendered objects. That improves object discoverability and consistency, but it also means repeated shared service-account names can still produce per-topic manifests with different label sets; this slice intentionally keeps the current per-topic artifact model.

## Out of scope

- Role/ClusterRole/RoleBinding generation
- imagePullSecrets, token secrets, or arbitrary ServiceAccount passthrough
- deduplicating or reconciling repeated ServiceAccount names across topics
