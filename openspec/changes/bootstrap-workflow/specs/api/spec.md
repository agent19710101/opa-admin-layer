## ADDED Requirements

### Requirement: validate tenant/topic admin specs
The product MUST reject invalid admin specs before rendering deployment plans.

#### Scenario: duplicate topic within tenant
- **WHEN** a tenant repeats the same topic name
- **THEN** validation returns an error

### Requirement: render OPA-only plans
The product MUST render a plan for each tenant/topic that includes the OPA bundle URL, OPA config YAML, and deployment YAML.

#### Scenario: render default bundle path
- **WHEN** a topic omits an explicit bundle resource
- **THEN** the system uses `<bundlePrefix>/<tenant>/<topic>.tar.gz`

#### Scenario: materialize rendered plan files
- **WHEN** the CLI render command is called with an output directory
- **THEN** the system writes `plan.json` plus per-tenant/topic `opa-config.yaml` and `deployment.yaml` files
