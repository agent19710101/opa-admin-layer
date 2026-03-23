## MODIFIED Requirements

### Requirement: render OPA-only plans
The product MUST render a plan for each tenant/topic that includes the OPA bundle URL, OPA config YAML, and deployment YAML.

#### Scenario: inherit shared Service metadata by default
- **WHEN** the input spec omits `topic.serviceType` and `topic.serviceAnnotations`
- **AND** shared `controlPlane.serviceType` and/or `controlPlane.serviceAnnotations` is configured
- **THEN** the rendered Service uses the shared Service type and annotation values

#### Scenario: override Service metadata per topic
- **WHEN** the input spec sets `topic.serviceType` and/or `topic.serviceAnnotations` for a tenant/topic
- **THEN** the rendered Service for that topic uses the topic Service type when provided
- **AND** topic annotation keys replace matching shared annotation keys while preserving shared annotations that were not overridden

#### Scenario: reject invalid topic Service metadata
- **WHEN** the input spec sets `topic.serviceType` to an unsupported Kubernetes Service type
- **OR** any `topic.serviceAnnotations` key is not a valid Kubernetes annotation key
- **THEN** validation fails before plan rendering