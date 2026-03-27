# Extracted requirements

## Workflow requirements

1. Use scheduler-driven automation only; do not rely on repo-local cron or host-specific timer config checked into the repo.
2. Persist project control-plane state outside the repository in a dedicated runtime state directory.
3. Run exactly one dispatcher slice per trigger.
4. Keep OpenSpec artifacts ahead of major implementation.
5. Persist changed files, validation results, sync state, current phase, and next scheduled action every run.

## Product requirements

1. Support OPA-only deployments first.
2. Treat tenant/topic scope as the unit of isolation and rollout.
3. Expose both CLI and REST API.
4. Validate input specs before rendering plans.
5. Render plan artifacts that are useful to operators: bundle URL, OPA config, config map manifest, deployment manifest, service manifest.
6. Generated deployment manifests must mount the generated OPA config instead of referencing an external file that is not provisioned by the plan.
7. Topic labels provided in the admin spec should propagate into generated Kubernetes manifests so operators can preserve ownership, environment, and routing metadata without post-processing.
8. Topic labels must be validated against Kubernetes label-key and label-value syntax before plan rendering so generated manifests remain valid.
9. Generated Kubernetes object names derived from spec, tenant, and topic identifiers must be validated before render so operators do not get unusable Deployment, ConfigMap, or Service artifacts.
10. Generated plans should expose each tenant/topic OPA workload through a stable Kubernetes Service derived from the normalized listen port.
11. Operators should be able to select the rendered Kubernetes Service type from the admin spec without post-processing manifests, while unsupported service modes are rejected during validation.
12. Operators should be able to attach shared Kubernetes Service annotations from the admin spec so common load-balancer and controller integrations do not require downstream patches.
13. Operators should be able to override shared Kubernetes Service type and annotations per topic while inheriting unspecified Service metadata from the shared control-plane defaults.
14. Operators should be able to declare shared and topic-level `externalTrafficPolicy` values for rendered Services when the effective Service type is `NodePort` or `LoadBalancer`, while invalid values or `ClusterIP` combinations are rejected during validation.
15. Operators should be able to declare shared and topic-level `sessionAffinity` values for rendered Services, limited to Kubernetes `None` and `ClientIP`, so sticky-client routing can be expressed without downstream patching.
16. Operators should be able to declare shared and topic-level `internalTrafficPolicy` values for rendered Services, limited to Kubernetes `Cluster` and `Local`, so in-cluster node-local routing can be expressed without downstream patching.
17. Operators should be able to declare shared OPA container CPU/memory requests and limits in the admin spec so rendered Deployments can carry baseline scheduling defaults without manual patching.
18. Operators should be able to override shared OPA CPU/memory requests and limits per topic while inheriting unspecified resource fields from the shared control-plane defaults.
19. Shared and inherited effective OPA resource profiles must reject CPU or memory requests that exceed their matching limits so generated Deployments stay within Kubernetes resource-budget rules.
20. Operators should be able to declare shared and topic-level `podAnnotations` values for rendered OPA pod templates so mesh, tracing, or admission-controller metadata can be expressed without downstream patches.
21. Operators should be able to declare shared and topic-level `deploymentAnnotations` values for rendered Deployment objects so rollout, ownership, or GitOps metadata can be expressed without downstream patches.
22. `controlPlane.baseServiceURL` must be validated as an absolute HTTP(S) URL before render so generated bundle URLs cannot be built from malformed or relative control-plane input.
23. `controlPlane.defaultListenAddress` must be validated as `:port`, `host:port`, or bracketed IPv6 `host:port` when provided so generated OPA args, Deployment probe ports, and Service ports cannot silently diverge.
24. Operators should be able to declare shared `controlPlane.replicas` values and topic-level `replicas` overrides so rendered Deployments can scale beyond a single replica without downstream patching.
25. Operators should be able to declare shared and topic-level `podLabels` values for rendered OPA pod templates so pod-only discovery, policy, or workload classification metadata can be expressed without mutating Services, ConfigMaps, or Deployment metadata.
26. Operators should be able to declare shared and topic-level `configMapAnnotations` values for rendered ConfigMaps so reloader, ownership, or GitOps metadata can be expressed without downstream patching, while keeping the slice limited to ConfigMap object annotations only.
27. Operators should be able to declare shared and topic-level `configMapLabels` values for rendered ConfigMaps so object-scoped discovery, GitOps, or policy labels can be expressed without mutating Services, Deployments, or pod templates.
28. Operators should be able to declare shared and topic-level `serviceAccountName` values for rendered OPA Deployments so workload identity and RBAC binding can be selected without downstream patches.
29. Operators should be able to declare shared and topic-level `automountServiceAccountToken` values for rendered OPA Deployments so service-account token projection can be explicitly enabled or disabled without downstream patches.
30. Operators should be able to declare shared and topic-level `imagePullPolicy` values for rendered OPA Deployments, limited to Kubernetes `Always`, `IfNotPresent`, and `Never`, so pull behavior can be controlled without downstream patches.
31. The repository should include runnable example admin specs for each supported primary input format so operator-facing ingestion paths are visible and easy to exercise.
32. Topics should be able to remove inherited Service, ConfigMap, Deployment, and pod annotation/label keys through explicit removal lists so shared metadata defaults can be cleared back to absent state without downstream patching.
33. Operators should be able to declare shared and topic-level `serviceAccountAnnotations` values, plus topic `removeServiceAccountAnnotations`, so rendered `ServiceAccount` objects can carry workload-identity metadata without downstream patching.
34. Operators should be able to declare shared and topic-level `serviceAccountLabels` values, plus topic `removeServiceAccountLabels`, so rendered `ServiceAccount` objects can carry ownership, GitOps, and policy labels without downstream patching.

## Operational requirements

1. Pin OPA versions in generated artifacts.
2. Allow the pinned OPA image reference to be overridden from the admin spec for environment-specific registry or tag policy.
3. Prefer sidecar-style deployment defaults.
4. Leave clear blockers when external capabilities are missing.
5. Keep the repository runnable and testable after each meaningful change.
32. Operators should be able to declare shared and topic-level `autoscaling` settings with CPU utilization targets so generated OPA workloads can emit HorizontalPodAutoscaler manifests without downstream patching.
33. Autoscaling-managed workloads must reject conflicting fixed `replicas` settings so the generated Deployment replica count and HPA controller contract cannot drift apart.
34. Autoscaling-managed workloads that target CPU utilization must require an effective `opaResources.requests.cpu` value after shared/topic inheritance so the generated HPA contract has a valid CPU-request baseline.
35. Autoscaling-managed workloads should also be able to target memory utilization, and any configured memory-utilization metric must require an effective inherited `opaResources.requests.memory` value after shared/topic inheritance.

36. Autoscaling behavior should also be able to express narrow scale-up/scale-down policy selection and scaling-step rules so generated HPAs can control step size without downstream patching.
