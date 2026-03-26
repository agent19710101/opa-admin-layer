# Documentation understanding summary

## Product goal

Build an OPA administration layer that helps operators define tenant/topic scoped policy distribution for OPA-only deployments, starting with a workflow-first repository that continuously ingests docs, plans with OpenSpec, implements small Go slices, validates, and prepares GitHub sync/release artifacts.

## What the documentation says

- OPA is the policy decision engine and should be kept close to the policy enforcement point for latency and resilience.
- OPA bundles and discovery are the built-in mechanisms for distributing policy and static data.
- OPAL adds realtime orchestration and client topic subscriptions, but it is not required for an OPA-only topology.
- Production deployments should pin versions, prefer TLS, and be explicit about bundle distribution and observability.
- Topic boundaries are the right mental model for routing policy/data updates to the right consumers.
- Topic-scoped metadata is operationally useful only if it survives plan rendering and reaches emitted deployment artifacts.
- Propagated Kubernetes metadata needs syntax validation at ingest time; otherwise the renderer can faithfully emit unusable manifests.
- Service exposure defaults should stay small, but operators still need a controlled way to switch between ClusterIP, NodePort, and LoadBalancer without forking generated YAML.
- Shared Service annotations are the next useful networking control after Service type selection; they let operators express controller/load-balancer metadata in the admin spec instead of patching generated manifests.
- Topics now inherit those Service defaults but can override `serviceType`, `externalTrafficPolicy`, `internalTrafficPolicy`, `sessionAffinity`, and merge `serviceAnnotations` when one workload needs different exposure or controller metadata than the rest of the fleet.
- Rendered OPA pod templates now follow the same inheritance pattern for `podAnnotations`, giving operators a narrow metadata escape hatch for mesh sidecar injection, tracing, or other admission-controller hints without widening into full deployment metadata control yet.
- Pod-template labels are now also first-class and separate from general topic labels: operators can set shared and topic-level `podLabels` for pod-only discovery, policy, or workload classification metadata without stamping those labels onto Services or ConfigMaps.
- Deployment-level annotations are the next small extension of that model: operators can now set shared and topic-level `deploymentAnnotations` for rollout, ownership, or GitOps metadata without opening arbitrary Deployment spec customization.
- Deployment-level labels are the next small workload-metadata extension after deployment annotations and pod-template labels: operators can now set shared and topic-level `deploymentLabels` for rollout tracking, ownership, or GitOps selectors without mutating Services, ConfigMaps, or pod templates.
- Service-account token projection now follows the same inherited workload-identity model as `serviceAccountName`: operators can set shared and topic-level `automountServiceAccountToken` values so rendered OPA Deployments can explicitly keep or disable projected API credentials without post-render patches.
- Generated ConfigMaps also need a narrow metadata escape hatch for reloader, ownership, or GitOps integrations; shared `controlPlane.configMapAnnotations` is the smallest useful contract because it covers ConfigMap-object metadata without widening into per-topic overrides or arbitrary ConfigMap shape changes.
- Shared Service traffic policy is now the next useful networking control after Service type and annotations: operators can express source-aware routing (`Cluster` vs `Local`) in the admin spec, while validation keeps that knob scoped to externally exposed Service types.
- Shared and topic-level Service internal traffic policy is the next narrow networking control after external traffic policy: operators can now express Kubernetes `Cluster` or `Local` in-cluster routing behavior in the admin spec with the same inheritance model used by other Service metadata.
- Shared and topic-level Service session affinity is the next narrow networking control after traffic policy: operators can now express `None` or `ClientIP` sticky-client behavior in the admin spec with the same inheritance model used by other Service metadata.
- Shared OPA resource defaults are a useful deployment baseline after image/probe/config/service realism; operators can now express baseline CPU/memory scheduling expectations once in the admin spec and selectively override them per topic when a noisy or latency-sensitive workload needs a different footprint.
- OPA resource validation now needs to cover effective inherited budgets, not just string syntax, because Kubernetes rejects CPU or memory requests that exceed their corresponding limits even when that mismatch only appears after topic overrides merge over shared defaults.
- Admin spec ingestion should match operator workflow realities by accepting both strict JSON and strict YAML through the same CLI and REST contract, and the repository should carry runnable examples for both formats so that support is visible to operators.
- `controlPlane.baseServiceURL` now follows the same up-front contract posture as Kubernetes-facing fields: validation requires an absolute HTTP(S) URL so bundle URL and OPA config rendering cannot silently normalize broken control-plane endpoints.
- `controlPlane.defaultListenAddress` is now validated before render when provided, with an intentionally small accepted contract (`:port`, `host:port`, or bracketed IPv6 `host:port`) so generated OPA args, probes, and Service ports stay aligned; the renderer now reuses the same strict parser instead of carrying a separate fallback port path.

## Locked project decisions

- default topology: OPA-only
- language: Go-only product code
- admin surface: CLI + REST API
- tenant model: multi-tenant topic-scoped shared control plane

## Immediate repository objective

Ship the workflow system first, then use it to ship the smallest useful vertical slice: validate an admin spec and render a runnable OPA deployment/config plan per tenant/topic, including the Kubernetes config map needed to mount the generated OPA config.

- Deployment scaling is now part of the spec contract through shared `controlPlane.replicas` defaults and topic-level `replicas` overrides.
