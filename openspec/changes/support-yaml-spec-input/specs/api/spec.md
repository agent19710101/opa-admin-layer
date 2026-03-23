## ADDED Requirements

### Requirement: accept JSON or YAML admin specs
The product MUST accept tenant/topic admin specifications encoded as either strict JSON or strict YAML through the shared ingestion path used by the CLI and REST API.

#### Scenario: validate YAML input
- **WHEN** an operator submits a valid admin spec encoded as YAML
- **THEN** validation succeeds using the same contract as JSON input

#### Scenario: reject unknown YAML fields
- **WHEN** an operator submits YAML containing a field that is not part of the admin spec contract
- **THEN** the product rejects the payload before rendering and reports an unknown-field style decode error
