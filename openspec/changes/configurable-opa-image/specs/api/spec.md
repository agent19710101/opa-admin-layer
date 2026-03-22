## MODIFIED Requirements

### Requirement: render OPA-only plans
The product MUST render a plan for each tenant/topic that includes the OPA bundle URL, OPA config YAML, and deployment YAML.

#### Scenario: render default OPA image
- **WHEN** the input spec omits `controlPlane.opaImage`
- **THEN** the rendered deployment uses the pinned default OPA image

#### Scenario: render configured OPA image
- **WHEN** the input spec sets `controlPlane.opaImage`
- **THEN** the rendered deployment uses that exact image reference
