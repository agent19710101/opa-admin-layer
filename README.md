# opa-admin-layer

OPA administration layer with a workflow-first delivery model.

## Product direction

Locked decisions:

- topology: OPA-only
- implementation language: Go-only for product code
- admin surface: CLI + REST API
- tenant model: multi-tenant topic-scoped

The workflow is the primary product. Application features are shipped as outputs of the workflow.

## Repository layout

- `cmd/opa-admin-layer`: CLI entrypoint
- `internal/admin`: tenant/topic scoped spec validation and plan rendering
- `internal/httpapi`: REST API for validation and plan generation
- `docs/understanding`: documentation-ingestion artifacts
- `deploy/examples`: runnable example admin specs
- `openspec`: change planning
- `scripts/cycle`: workflow scripts
- `scripts/automation`: thin phase runner wrapper
- `.github/workflows`: CI and release preparation

## Runtime state

Runtime state is intentionally stored outside the repository.

Defaults:

- state directory: `${XDG_STATE_HOME:-$HOME/.local/state}/opa-admin-layer`
- override: set `OPA_ADMIN_LAYER_STATE_DIR`

The runtime state directory holds the active project, project queue, switch request, cycle status, summary, blockers, and logs.

## Quick start

```bash
make test
./bin/opa-admin-layer validate -input deploy/examples/dev-spec.json
./bin/opa-admin-layer render -input deploy/examples/dev-spec.json
./bin/opa-admin-layer serve -addr :8080
```

Build first:

```bash
make build
```

Example API usage:

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/v1/validate \
  -H 'content-type: application/json' \
  --data @deploy/examples/dev-spec.json
curl -s http://localhost:8080/v1/plans \
  -H 'content-type: application/json' \
  --data @deploy/examples/dev-spec.json
```

## Workflow loop

A scheduler should trigger exactly three jobs during the active window:

- `dev-preflight` at `55 19 * * *` Europe/Warsaw
- `dev-cycle-dispatch` at `*/15 20-23,0-4 * * *` Europe/Warsaw
- `dev-closeout` at `0 5 * * *` Europe/Warsaw

Use `scripts/automation/run-phase.sh` as the repo entrypoint.

The dispatcher advances one phase per run in this order:

1. preflight and control-plane check
2. documentation ingestion
3. architecture/repo layout
4. OpenSpec update
5. smallest useful implementation slice
6. verification
7. sync and release preparation
8. persist cycle state

## Current vertical slice

The first shipped slice validates a tenant/topic scoped admin spec and renders an OPA-only plan containing:

- normalized tenant/topic inventory
- per-topic OPA bundle URL
- generated OPA config YAML
- generated Kubernetes deployment YAML for a sidecar-style OPA deployment

This slice is exposed through both the CLI and the REST API.
