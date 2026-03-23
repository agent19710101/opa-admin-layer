# Changelog

## Unreleased

- add shared/topic `internalTrafficPolicy` support for rendered Services with inherited `Cluster`/`Local` rendering
- add shared/topic `sessionAffinity` support for rendered Services with inherited `None`/`ClientIP` rendering
- add shared/topic `externalTrafficPolicy` support for rendered Services with effective compatibility validation
- enforce shared and inherited OPA resource request/limit budgets during validation
- add topic-level `serviceType` and `serviceAnnotations` overrides merged over shared Service defaults
- add shared `controlPlane.opaResources` support for rendered OPA Deployment CPU/memory requests and limits
- bootstrap repository workflow, documentation ingestion artifacts, OpenSpec scaffolding, cycle scripts, and first Go vertical slice for tenant/topic scoped OPA plan rendering
