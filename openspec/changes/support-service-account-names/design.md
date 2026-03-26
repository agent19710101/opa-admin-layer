# Design: shared and topic service account names

## Scope

This slice adds one narrow workload-identity knob to rendered OPA Deployments:

- shared `controlPlane.serviceAccountName`
- topic-level `serviceAccountName` overrides
- validation using Kubernetes DNS-subdomain name rules
- rendering only into `Deployment.spec.template.spec.serviceAccountName`
- inheritance where topic values replace, rather than merge with, the shared default

## Rationale

Service account choice is one of the most common Kubernetes workload integration points because it controls RBAC identity, projected cloud credentials, and admission-policy targeting. It belongs next to other workload-level defaults the renderer already owns, and it is materially safer than opening arbitrary pod-spec passthrough.

The first slice keeps the contract string-based and intentionally simple. A topic can opt into a different service account, but there is no deletion semantic yet for clearing an inherited shared value back to the namespace default. That keeps validation and rendering straightforward while still covering the common operator path.

## Out of scope

- service account creation or RBAC generation
- deletion semantics for removing an inherited shared service account name
- arbitrary pod spec customization such as automount toggles, tolerations, or affinity
