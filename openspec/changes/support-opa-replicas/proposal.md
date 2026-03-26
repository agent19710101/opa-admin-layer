# Proposal: support configurable OPA replicas

## Why

Rendered Deployments still hard-code `spec.replicas: 1`. That keeps the first slice simple, but it forces operators to patch generated manifests whenever they need extra read capacity or basic redundancy.

## What changes

- add shared `controlPlane.replicas` to the admin spec
- add topic-level `replicas` overrides for tenant/topic workloads
- validate that replica counts are zero-or-greater in raw input and normalize zero to inherited/default behavior
- render the effective replica count into generated Deployment manifests
- document the inheritance/defaulting contract in repo docs and examples

## Scope

This slice stays intentionally narrow:

- only `Deployment.spec.replicas` becomes configurable
- no autoscaling/HPA objects
- no surge/rolling-update strategy controls yet
