# Missing decisions

The core product direction is now explicit. Remaining design choices are implementation-level rather than product-blocking.

## Open questions to revisit later

- bundle publication backend: static object storage, GitHub Releases, or dedicated control-plane endpoint
- authentication/authorization model for the REST API
- exact persistence model for workflow telemetry and historical runs
- whether the admin spec should also support YAML in addition to JSON
- whether future rendered Services need configurable type/annotations for ingress, mesh, or load-balancer integration
