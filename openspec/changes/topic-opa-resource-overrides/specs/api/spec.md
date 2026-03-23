## MODIFIED Requirements

### Requirement: render OPA-only plans
The product MUST render a plan for each tenant/topic that includes the OPA bundle URL, OPA config YAML, and deployment YAML.

#### Scenario: inherit shared OPA deployment resources by default
- **WHEN** the input spec omits `topic.opaResources`
- **AND** shared `controlPlane.opaResources` is configured
- **THEN** the rendered deployment uses the shared resource values

#### Scenario: override shared OPA deployment resources per topic
- **WHEN** the input spec sets `topic.opaResources` for a tenant/topic
- **THEN** the rendered deployment includes the topic override values for the matching requests/limits fields
- **AND** any unspecified resource fields continue to inherit from shared `controlPlane.opaResources`

#### Scenario: reject empty or invalid topic OPA resource sections
- **WHEN** the input spec sets `topic.opaResources.requests` or `topic.opaResources.limits` without at least one of `cpu` or `memory`
- **OR** a topic-level CPU or memory value is not a valid Kubernetes quantity
- **THEN** validation fails before plan rendering
