# Missing decisions

The core product direction is now explicit. Remaining design choices are implementation-level rather than product-blocking.

## Open questions to revisit later

- bundle publication backend: static object storage, GitHub Releases, or dedicated control-plane endpoint
- authentication/authorization model for the REST API
- exact persistence model for workflow telemetry and historical runs
- whether namespace control should later expand beyond shared `controlPlane.namespace` into per-topic overrides or namespace creation/ownership policy
- whether Service metadata control should later expand beyond shared/topic `serviceType`, `externalTrafficPolicy`, `internalTrafficPolicy`, `sessionAffinity`, `serviceAnnotations`, and `serviceLabels` into multi-port exposure or deeper ingress/mesh/load-balancer integration
- whether workload control should later expand beyond shared/topic `replicas`, shared/topic `autoscaling` (now with CPU/memory utilization targets, optional stabilization windows, and effective request requirements for configured utilization metrics), rendered `ServiceAccount` manifests for uniquely owned effective `serviceAccountName` values, shared/topic `serviceAccountAnnotations` plus topic `removeServiceAccountAnnotations`, shared/topic `serviceAccountLabels` plus topic `removeServiceAccountLabels`, shared/topic `configMapAnnotations`, shared/topic `configMapLabels`, shared/topic `serviceAccountName`, shared/topic `imagePullPolicy`, shared/topic `automountServiceAccountToken`, `deploymentAnnotations`, deployment `deploymentLabels`, pod-template `podAnnotations`, and pod-template `podLabels` into custom/external HPA metrics, ServiceAccount RBAC generation, a future shared-object model for intentionally reused `serviceAccountName` values beyond today's binding-only suppression, or other admission-controller integration knobs
- whether OPA resource policy should later expand beyond the new request<=limit guardrails into allowed ranges, ratio checks, or admission-style budgeting rules
