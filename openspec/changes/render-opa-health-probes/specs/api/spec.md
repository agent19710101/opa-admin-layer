## MODIFIED Requirements

### Requirement: render OPA-only plans
The product MUST render a plan for each tenant/topic that includes the OPA bundle URL, OPA config YAML, and deployment YAML.

#### Scenario: render deployment container port from default listen address
- **WHEN** the input spec omits `controlPlane.defaultListenAddress`
- **THEN** the rendered deployment declares container port `8181`
- **AND** the OPA container liveness probe targets `/health`
- **AND** the OPA container readiness probe targets `/health?plugins`

#### Scenario: render deployment container port from explicit listen address
- **WHEN** the input spec sets `controlPlane.defaultListenAddress` to a specific host:port value
- **THEN** the rendered deployment declares the parsed port for the OPA container
- **AND** both health probes use that same declared container port