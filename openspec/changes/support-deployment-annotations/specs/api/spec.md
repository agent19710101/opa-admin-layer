## ADDED Requirements

### Requirement: Shared deployment annotations
The admin spec SHALL allow an optional `controlPlane.deploymentAnnotations` map whose keys must satisfy the Kubernetes metadata-key contract.

#### Scenario: Shared deployment annotations validate and render
- **WHEN** a spec sets `controlPlane.deploymentAnnotations`
- **THEN** validation accepts Kubernetes-valid annotation keys
- **AND** generated Deployment manifests render the annotations under `metadata.annotations`

### Requirement: Topic deployment annotation overrides
The admin spec SHALL allow a topic to set optional `deploymentAnnotations` that merge over shared control-plane deployment annotations key-by-key.

#### Scenario: Topic deployment annotations override shared keys
- **GIVEN** shared `controlPlane.deploymentAnnotations`
- **WHEN** a topic sets `deploymentAnnotations`
- **THEN** the rendered Deployment metadata contains the union of both maps
- **AND** topic-provided values replace shared values for matching keys
