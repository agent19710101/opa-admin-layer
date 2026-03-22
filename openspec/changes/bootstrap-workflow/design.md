# Design

## Workflow design

The workflow is controlled by persisted JSON state instead of cron phase fan-out. A scheduler only wakes three isolated jobs. The dispatcher reads the runtime state file, executes exactly one phase, records the result, and advances the pointer.

## Product design

The first slice models a tenant/topic scoped admin spec. Validation enforces uniqueness and required control-plane inputs. Rendering applies defaults and produces a simple OPA-only plan with pinned OPA image, bundle URL, config YAML, and deployment YAML.

## Why this slice first

It is small, testable, useful to operators, and aligned with the documented OPA-only starting topology. It also creates a clean seam for later features such as bundle publishing, signing, API auth, and release-driven rollout.
