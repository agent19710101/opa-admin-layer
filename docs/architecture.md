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
