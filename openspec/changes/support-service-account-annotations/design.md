# Design: rendered ServiceAccount annotations

## Scope

This slice extends the existing workload-identity contract just enough to let generated ServiceAccounts carry object-scoped annotation metadata.

- reuse the existing shared/topic inheritance pattern already used for Service, ConfigMap, Deployment, and pod annotations
- allow topics to override shared keys and explicitly remove inherited keys through `removeServiceAccountAnnotations`
- render annotations only into `ServiceAccount.metadata.annotations`

## Rationale

ServiceAccount annotations are the smallest useful next workload-identity step after rendering `serviceaccount.yaml` because common cloud-identity integrations depend on annotations rather than labels or broader RBAC objects.

## Tradeoffs

This slice keeps ServiceAccount metadata asymmetric for now: annotations are supported, but labels and other ServiceAccount fields remain out of scope. That bias is intentional because annotations solve the most common workload-identity integration need while keeping the contract narrow.

## Out of scope

- ServiceAccount labels or removal lists for labels
- Role/ClusterRole/RoleBinding generation
- imagePullSecrets, token secrets, or arbitrary ServiceAccount passthrough
