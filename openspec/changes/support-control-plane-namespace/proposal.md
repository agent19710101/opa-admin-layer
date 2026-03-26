# Proposal: support shared control-plane namespace

## Why

The renderer now owns full ConfigMap, Deployment, and Service manifests for each tenant/topic workload, but it still assumes the Kubernetes default namespace. Operators that deploy into a dedicated namespace must patch every generated manifest after render even though namespace placement is a shared control-plane concern.

## What changes

- add optional `controlPlane.namespace` to the admin spec
- validate the field against Kubernetes namespace-compatible DNS label syntax
- render the configured namespace into generated ConfigMap, Deployment, and Service manifests
- keep the first slice intentionally shared-only; per-topic namespace overrides remain out of scope

## Impact

Operators can place the entire rendered OPA workload set into one explicit namespace without downstream manifest edits, while the contract stays narrow enough to avoid early namespace-ownership and multi-namespace policy complexity.
