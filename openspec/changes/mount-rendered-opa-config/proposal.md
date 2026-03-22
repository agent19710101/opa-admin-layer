# Proposal: mount rendered OPA config

## Why

The rendered deployment manifest points OPA at `/config/opa-config.yaml`, but the current plan only emits a standalone config file alongside the deployment YAML. Operators must manually bridge that gap before the workload is runnable.

## Change

- emit a Kubernetes ConfigMap manifest per tenant/topic containing `opa-config.yaml`
- mount that ConfigMap in the rendered Deployment at `/config`
- export the ConfigMap manifest in the plan tree and cover the new behavior with tests

## Impact

The generated plan becomes self-contained for the config path it already advertises. Operators can apply the rendered manifests directly with less manual glue code and lower risk of configuration drift.
