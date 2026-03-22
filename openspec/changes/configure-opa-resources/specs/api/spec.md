## MODIFIED Requirements

### Requirement: render OPA-only plans
The product MUST render a plan for each tenant/topic that includes the OPA bundle URL, OPA config YAML, and deployment YAML.

#### Scenario: omit deployment resources by default
- **WHEN** the input spec omits `controlPlane.opaResources`
- **THEN** the rendered deployment omits a `resources` block

#### Scenario: render configured OPA deployment resources
- **WHEN** the input spec sets `controlPlane.opaResources.requests` and/or `controlPlane.opaResources.limits`
- **THEN** the rendered deployment includes a `resources` block for the OPA container
- **AND** configured CPU and memory values are emitted under the matching requests/limits section

#### Scenario: reject empty OPA resource sections
- **WHEN** the input spec sets `controlPlane.opaResources.requests` or `controlPlane.opaResources.limits` without at least one of `cpu` or `memory`
- **THEN** validation fails before plan rendering
