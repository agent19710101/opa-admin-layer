# Changelog

## Unreleased

- add shared `controlPlane.configMapAnnotations` support for rendered ConfigMap metadata with Kubernetes annotation-key validation
- add shared/topic `podLabels` support for rendered Deployment pod templates with selector-safe inherited label merging

- add shared `controlPlane.replicas` plus topic-level `replicas` overrides for rendered Deployment scaling
- add shared/topic `deploymentAnnotations` support for rendered Deployment metadata with inherited annotation merging
- add shared/topic `podAnnotations` support for rendered OPA pod templates with inherited annotation merging
- add shared `controlPlane.namespace` support with Kubernetes namespace validation and namespace rendering in generated ConfigMap, Deployment, and Service manifests
- validate REST request `Content-Type` for `/v1/validate` and `/v1/plans`, accepting JSON/YAML media types and returning `415 Unsupported Media Type` for unsupported payload types
- add shared/topic `internalTrafficPolicy` support for rendered Services with inherited `Cluster`/`Local` rendering
- add shared/topic `sessionAffinity` support for rendered Services with inherited `None`/`ClientIP` rendering
- add shared/topic `externalTrafficPolicy` support for rendered Services with effective compatibility validation
- enforce shared and inherited OPA resource request/limit budgets during validation
- add topic-level `serviceType` and `serviceAnnotations` overrides merged over shared Service defaults
- add shared `controlPlane.opaResources` support for rendered OPA Deployment CPU/memory requests and limits
- bootstrap repository workflow, documentation ingestion artifacts, OpenSpec scaffolding, cycle scripts, and first Go vertical slice for tenant/topic scoped OPA plan rendering
