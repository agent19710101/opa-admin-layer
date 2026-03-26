# Architecture

## Control plane

The project has a single active-project model. Runtime control files live outside the repository in the configured state directory. Future project switches must flow through the switch-request file and be handled by `scripts/cycle/preflight.sh`.

## Workflow plane

A scheduler triggers three entrypoints:

- `preflight.sh`: validate control-plane state and repo health before the active window
- `dispatch.sh`: execute exactly one phase slice from the ordered workflow sequence
- `closeout.sh`: summarize nightly state and verify the repo before the inactive window

Dispatcher phase order:

1. preflight
2. ingest
3. architecture
4. openspec
5. implement
6. verify
7. sync
8. persist

## Product plane

The first product slice has three layers:

- `internal/admin`: specification model, validation, normalization, and OPA plan rendering
- `internal/httpapi`: REST surface for validation and plan generation
- `cmd/opa-admin-layer`: CLI surface for render, validate, and serve

The output plan is intentionally small and operator-focused:

- normalized tenant/topic inventory
- bundle URL per topic
- rendered OPA config YAML
- rendered Kubernetes deployment YAML

Current architecture note: deployment rendering now treats the OPA image as control-plane configuration instead of a hard-coded renderer constant. The default remains pinned, but operators can override the image reference through the admin spec to satisfy registry and release policy without post-processing manifests.

Architecture update (2026-03-22, config mount slice): each tenant/topic render now emits two Kubernetes artifacts that intentionally pair together — a ConfigMap containing `opa-config.yaml` and a Deployment that mounts that ConfigMap at `/config`. This keeps the generated plan self-contained and removes an integration gap where the Deployment referenced a file the renderer did not provision.

Architecture update (2026-03-22, topic label propagation): topic labels are now treated as operator metadata that should flow through render output instead of remaining plan-only data. The renderer keeps a small built-in identity label set for selection and ownership, then merges topic labels into ConfigMap metadata, Deployment metadata, and pod-template labels without letting user labels override the built-in identity keys.

Architecture update (2026-03-22, deployment health probes): rendered Deployments now derive an OPA container port from the normalized listen address and use that same port for default readiness and liveness probes. This keeps the existing single-container shape while making rollout readiness and restart signaling part of the generated plan instead of downstream patch work.

Architecture update (2026-03-22, service manifest rendering): each tenant/topic render now emits a Kubernetes Service that targets the generated OPA Deployment on the same derived HTTP port. This keeps the plan self-contained for in-cluster reachability without expanding the spec surface with early service-type or ingress decisions.

Architecture update (2026-03-22, topic label validation): topic labels remain the right operator metadata escape hatch, but they now enter the system through a stricter contract. The validation layer rejects Kubernetes-invalid label keys and values before render so propagated labels cannot silently poison ConfigMap, Deployment, or Service manifests.

Architecture update (2026-03-22, rendered resource-name validation): generated workload object names are now treated as part of the validation contract instead of a renderer implementation detail. Because the product derives Deployment, ConfigMap, and Service names from spec/tenant/topic identifiers and also reuses the deployment name in built-in labels, the validation layer now rejects combinations that would exceed the safe DNS-1123/label budget before any manifests are emitted.

Architecture update (2026-03-22, service type configurability): the generated Service remains intentionally minimal, but service exposure is no longer hard-coded to ClusterIP. The shared control-plane spec can now request `ClusterIP`, `NodePort`, or `LoadBalancer`, with validation keeping unsupported Kubernetes service modes out of both CLI and REST paths.

Architecture update (2026-03-23, service annotations): generated Services now accept a shared control-plane `serviceAnnotations` map so controller-specific metadata can travel with the Service manifest the renderer already owns. The scope stays intentionally narrow to Service metadata only, and annotation keys are validated before render while values are YAML-quoted in output so common health-check and load-balancer strings remain safe to emit.

Architecture update (2026-03-23, OPA resource defaults): generated Deployments now accept a shared control-plane `opaResources` block for CPU/memory requests and limits. The scope stays intentionally small and workload-wide: one default OPA container resource profile fans out to every rendered tenant/topic Deployment, giving operators a scheduling baseline without introducing deeper policy knobs.

Architecture update (2026-03-23, topic OPA resource overrides): tenant/topic entries can now carry their own `opaResources` block that merges over the shared control-plane defaults field-by-field. This keeps the operator contract compact: shared defaults still define the baseline, while a topic only needs to specify the CPU or memory value it wants to replace instead of restating the entire resource profile.

Architecture update (2026-03-23, OPA resource quantity validation): shared `controlPlane.opaResources` values and topic-level `opaResources` overrides are now treated as part of the admin contract rather than opaque manifest text. The validation layer parses CPU and memory quantities before render so malformed values fail in CLI and REST validation paths instead of leaking into generated Deployments and only failing at Kubernetes apply time.

Architecture update (2026-03-23, OPA resource budget guardrails): effective OPA resource profiles are now validated after inheritance, not just syntax-checked as raw strings. Shared control-plane requests cannot exceed their matching limits, and topic-level overrides are rechecked after merging over shared defaults so one topic cannot accidentally render a Deployment with request values above the inherited CPU or memory limit.

Architecture update (2026-03-23, topic Service overrides): Service exposure and annotation metadata now follow the same inheritance model as topic OPA resources. Shared control-plane Service defaults still define the baseline, but a topic can now override `serviceType` and merge `serviceAnnotations` key-by-key for workload-specific ingress, load-balancer, or mesh integration without downstream manifest patching.

Architecture update (2026-03-23, external traffic policy): rendered Services now treat `externalTrafficPolicy` as inherited Service metadata alongside `serviceType` and annotations. The control plane can set a shared `Cluster` or `Local` policy, topics can override it, and validation rechecks the effective combination so source-aware traffic policy is only emitted when the final Service type is `NodePort` or `LoadBalancer`.

Architecture update (2026-03-23, service session affinity): rendered Services now treat `sessionAffinity` as inherited Service metadata alongside `serviceType`, traffic policy, and annotations. The control plane can set a shared `None` or `ClientIP` default, topics can override it, and validation keeps the first slice intentionally narrow to Kubernetes-supported affinity modes without introducing deeper session-affinity config yet.

Architecture update (2026-03-23, internal traffic policy): rendered Services now treat `internalTrafficPolicy` as inherited Service metadata alongside `serviceType`, traffic policy, session affinity, and annotations. The control plane can set a shared `Cluster` or `Local` default, topics can override it, and validation keeps the slice intentionally narrow to Kubernetes-supported in-cluster routing modes without widening into port or ingress configuration.

Architecture update (2026-03-23, YAML spec ingestion): spec decoding is now format-flexible but contract-strict. A single shared ingestion path accepts either JSON or YAML for CLI and REST flows, while unknown fields are still rejected before validation/render so YAML support does not reopen the loose-schema drift that the strict JSON decoder previously closed.

Architecture update (2026-03-26, REST content-type contract): the REST API now treats request media type as part of the operator contract instead of silently attempting to decode any posted body. `/v1/validate` and `/v1/plans` accept `application/json`, `application/yaml`, `application/x-yaml`, `text/yaml`, `text/x-yaml`, or an empty `Content-Type` for simple callers, and they return `415 Unsupported Media Type` for other request types.

Architecture update (2026-03-26, shared Kubernetes namespace): rendered workload placement is now a shared control-plane concern instead of an implicit default-namespace assumption. `controlPlane.namespace` optionally fans out to every generated ConfigMap, Deployment, and Service manifest, while validation keeps the first slice narrow by requiring a Kubernetes namespace-compatible DNS label and leaving per-topic namespace overrides out of scope.

Architecture update (2026-03-26, pod template annotations): rendered Deployments now treat pod-template annotations as inherited workload metadata alongside Service annotations and OPA resources. The control plane can set shared `podAnnotations`, topics can override or extend them key-by-key, and validation keeps the slice narrow by checking annotation keys with the existing Kubernetes metadata-key contract while leaving deployment-level annotations out of scope for now.

Architecture update (2026-03-26, deployment annotations): rendered Deployments now also treat top-level Deployment annotations as inherited workload metadata. The control plane can set shared `deploymentAnnotations`, topics can override or extend them key-by-key, and rendering keeps the surface intentionally narrow by writing only `Deployment.metadata.annotations` instead of opening arbitrary Deployment customization.

Architecture update (2026-03-23, checked-in YAML example): the repository now carries a first-class YAML example spec alongside the JSON example. That keeps the supported operator input path visible and testable in-tree instead of leaving YAML as an implementation detail only covered by unit tests and README snippets.

Architecture update (2026-03-23, control-plane URL validation): `controlPlane.baseServiceURL` is now treated as a first-class endpoint contract rather than a non-empty string. Validation requires an absolute HTTP(S) URL with a host and no fragment so bundle URL composition and rendered OPA config cannot be built from malformed control-plane input.

Architecture update (2026-03-23, listen-address validation): `controlPlane.defaultListenAddress` is now treated as a shared socket contract instead of best-effort renderer input. Validation accepts only `:port`, `host:port`, or bracketed IPv6 `host:port` forms so the rendered OPA `--addr` argument, Deployment port/probe wiring, and Service target port all derive from the same parseable value. The renderer now reuses that same parser directly, removing the old fallback path that silently defaulted ports to `8181` when the address string could not be parsed.

Architecture update (2026-03-26, configurable replicas): rendered Deployments no longer hard-code one replica. The control plane can set a shared `replicas` default, topics can override it, validation keeps raw values non-negative, and normalization preserves the existing single-replica behavior when neither scope sets a value.
