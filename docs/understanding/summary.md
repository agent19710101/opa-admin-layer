# Documentation understanding summary

## Product goal

Build an OPA administration layer that helps operators define tenant/topic scoped policy distribution for OPA-only deployments, starting with a workflow-first repository that continuously ingests docs, plans with OpenSpec, implements small Go slices, validates, and prepares GitHub sync/release artifacts.

## What the documentation says

- OPA is the policy decision engine and should be kept close to the policy enforcement point for latency and resilience.
- OPA bundles and discovery are the built-in mechanisms for distributing policy and static data.
- OPAL adds realtime orchestration and client topic subscriptions, but it is not required for an OPA-only topology.
- Production deployments should pin versions, prefer TLS, and be explicit about bundle distribution and observability.
- Topic boundaries are the right mental model for routing policy/data updates to the right consumers.

## Locked project decisions

- default topology: OPA-only
- language: Go-only product code
- admin surface: CLI + REST API
- tenant model: multi-tenant topic-scoped shared control plane

## Immediate repository objective

Ship the workflow system first, then use it to ship the smallest useful vertical slice: validate an admin spec and render an OPA deployment/config plan per tenant/topic.
