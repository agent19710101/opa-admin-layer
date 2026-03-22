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
5. Render plan artifacts that are useful to operators: bundle URL, OPA config, deployment manifest.

## Operational requirements

1. Pin OPA versions in generated artifacts.
2. Allow the pinned OPA image reference to be overridden from the admin spec for environment-specific registry or tag policy.
3. Prefer sidecar-style deployment defaults.
4. Leave clear blockers when external capabilities are missing.
5. Keep the repository runnable and testable after each meaningful change.
