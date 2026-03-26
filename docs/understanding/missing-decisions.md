# Missing decisions

The core product direction is now explicit. Remaining design choices are implementation-level rather than product-blocking.

## Open questions to revisit later

- bundle publication backend: static object storage, GitHub Releases, or dedicated control-plane endpoint
- authentication/authorization model for the REST API
- exact persistence model for workflow telemetry and historical runs
- whether namespace control should later expand beyond shared `controlPlane.namespace` into per-topic overrides or namespace creation/ownership policy
- whether Service metadata control should later expand beyond shared/topic `serviceType`, `externalTrafficPolicy`, `internalTrafficPolicy`, `sessionAffinity`, and `serviceAnnotations` into multi-port exposure or deeper ingress/mesh/load-balancer integration
- whether OPA resource policy should later expand beyond the new request<=limit guardrails into allowed ranges, ratio checks, or admission-style budgeting rules
