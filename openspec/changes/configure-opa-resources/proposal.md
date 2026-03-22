# Proposal: configure OPA deployment resources

## Why

Generated Deployments now cover config mounting, health probes, and Service exposure, but they still omit container resource requests and limits. That keeps rendered workloads runnable, yet it pushes a common production hardening step back onto operators and weakens scheduler realism for the manifests this project owns.

## Change

- add an optional shared `controlPlane.opaResources` block to the admin spec
- support CPU and memory values under `requests` and `limits`
- render the configured values into every generated OPA Deployment manifest
- reject empty `requests` or `limits` objects so the new field stays intentional

## Impact

Operators can express baseline OPA scheduling defaults once in the admin spec instead of patching every rendered Deployment downstream. The scope stays small and shared across tenant/topic workloads, leaving quantity validation and per-topic overrides for future slices.
