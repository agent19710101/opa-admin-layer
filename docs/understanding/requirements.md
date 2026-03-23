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
13. Operators should be able to declare shared OPA container CPU/memory requests and limits in the admin spec so rendered Deployments can carry baseline scheduling defaults without manual patching.
14. `controlPlane.baseServiceURL` should be validated as an absolute HTTP(S) URL before render so generated bundle URLs cannot be built from malformed or relative control-plane input.

## Operational requirements

1. Pin OPA versions in generated artifacts.
2. Allow the pinned OPA image reference to be overridden from the admin spec for environment-specific registry or tag policy.
3. Prefer sidecar-style deployment defaults.
4. Leave clear blockers when external capabilities are missing.
5. Keep the repository runnable and testable after each meaningful change.
