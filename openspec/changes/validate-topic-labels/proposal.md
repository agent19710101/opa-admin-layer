# Proposal: validate propagated topic labels

## Why

Topic labels now flow directly into rendered ConfigMap, Deployment, and Service manifests. Without validating Kubernetes label syntax at spec-ingest time, a single malformed label key or value can make the generated plan invalid and push a debugging loop onto operators.

## Change

- validate topic label keys against Kubernetes label-key syntax before plan rendering
- validate topic label values against Kubernetes label-value syntax before plan rendering
- return the same validation failures through the CLI and REST API surfaces
- add regression tests and documentation for the stricter contract

## Impact

Rendered plans stay operator-usable by rejecting malformed metadata early, before invalid manifests are emitted into the plan tree.
