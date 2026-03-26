## ADDED Requirements

### Requirement: accept only supported REST admin-spec media types
The REST API MUST accept strict JSON or YAML admin-spec payloads only when the request `Content-Type` is empty or identifies a supported JSON/YAML media type.

#### Scenario: validate JSON with explicit media type
- **WHEN** an operator submits a valid admin spec to `/v1/validate` with `Content-Type: application/json`
- **THEN** validation succeeds

#### Scenario: validate YAML with explicit media type
- **WHEN** an operator submits a valid admin spec to `/v1/validate` with `Content-Type: application/yaml`
- **THEN** validation succeeds

#### Scenario: reject unsupported media type
- **WHEN** an operator submits an admin spec to `/v1/validate` or `/v1/plans` with an unsupported `Content-Type`
- **THEN** the API returns `415 Unsupported Media Type` and explains which media types are accepted

#### Scenario: accept missing media type
- **WHEN** an operator submits a valid admin spec without a `Content-Type` header
- **THEN** the API still decodes the payload using the existing JSON/YAML detection path
