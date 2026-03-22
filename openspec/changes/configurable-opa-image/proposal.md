# Proposal: configurable OPA image

## Why

The current plan renderer hard-codes the OPA container image in generated deployment manifests. That keeps the first slice simple, but it prevents operators from pinning an approved image tag or registry mirror per environment without patching generated output after the fact.

## Change

- add an optional `controlPlane.opaImage` field to the admin spec
- default the field to the current pinned image when omitted
- render the configured image into generated deployment manifests
- add regression tests for defaulted and overridden image behavior

## Impact

Operators can keep generated manifests aligned with environment-specific image policy while preserving the current secure default for existing specs.
