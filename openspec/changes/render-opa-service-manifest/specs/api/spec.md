## MODIFIED Requirements

### Requirement: render OPA-only plans
The product MUST render a plan for each tenant/topic that includes the OPA bundle URL, OPA config YAML, deployment YAML, and Service YAML.

#### Scenario: render service manifest from derived listen port
- **WHEN** a tenant/topic plan is rendered
- **THEN** the output includes a Kubernetes Service manifest
- **AND** the Service port matches the OPA port derived from the normalized listen address
- **AND** the Service targetPort routes to the generated OPA container port

#### Scenario: materialize rendered plan files
- **WHEN** the CLI render command is called with an output directory
- **THEN** the system writes `plan.json` plus per-tenant/topic `opa-config.yaml`, `configmap.yaml`, `deployment.yaml`, and `service.yaml` files
