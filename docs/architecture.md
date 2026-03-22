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
