# Changelog

## Unreleased

- extend autoscaling to support optional memory utilization targets and require effective inherited CPU/memory requests for configured utilization metrics
- add narrow HPA behavior support for scale-up and scale-down stabilization windows with validation and manifest rendering
- validate autoscaling workloads require an effective `opaResources.requests.cpu` value when using CPU utilization targets
- add topic-level inherited metadata removal lists so shared Service, ConfigMap, Deployment, and pod annotation/label keys can be cleared back to absent state without downstream patching
- add shared/topic `serviceLabels` support for rendering Service-only metadata labels with validation and immutable built-in identity keys

- add shared/topic `configMapAnnotations` support for rendered ConfigMap metadata with inherited annotation merging and Kubernetes annotation-key validation
- add shared/topic `configMapLabels` support for rendered ConfigMap metadata with inherited label merging and selector-safe built-in label protection
- add shared/topic `podLabels` support for rendered Deployment pod templates with selector-safe inherited label merging
- render `ServiceAccount` manifests and `serviceaccount.yaml` artifacts whenever workloads resolve an explicit effective `serviceAccountName`
- add shared/topic `serviceAccountName` support for rendered Deployment workload identity binding
- add shared/topic `automountServiceAccountToken` support for rendered Deployment service-account token projection control

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
