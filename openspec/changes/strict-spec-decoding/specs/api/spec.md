## MODIFIED Requirements

### Requirement: validate tenant/topic admin specs
The product MUST reject invalid admin specs before rendering deployment plans, including malformed JSON and unsupported JSON fields.

#### Scenario: unknown field in CLI input
- **WHEN** a spec file includes an unsupported JSON field
- **THEN** the CLI returns a decoding error before plan rendering

#### Scenario: unknown field in REST input
- **WHEN** a REST request body includes an unsupported JSON field
- **THEN** the API returns `400 Bad Request` with a decoding error before plan rendering
