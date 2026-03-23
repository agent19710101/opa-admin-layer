# Proposal: validate OPA resource quantity syntax

## Why

The renderer now accepts shared OPA CPU/memory requests and limits, but it still treats those values as opaque strings. That means a spec can pass validation while still emitting Deployment manifests with Kubernetes-invalid quantities, pushing a basic configuration error into apply time instead of catching it in the admin layer.

## Change

- validate `controlPlane.opaResources` CPU and memory strings with Kubernetes quantity parsing before render
- return field-specific validation errors for invalid request and limit values
- add regression coverage for CLI, admin validation, and REST validation paths
- document the stricter contract in README and design notes

## Impact

Operators get earlier feedback for malformed resource values, and generated Deployment manifests better reflect a contract the cluster can actually accept.
