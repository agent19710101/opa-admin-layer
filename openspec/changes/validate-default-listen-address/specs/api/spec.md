## MODIFIED Requirements

### Requirement: validate tenant/topic admin specs
The product MUST reject invalid admin specs before rendering deployment plans, including malformed control-plane listen addresses that would make rendered OPA socket configuration disagree with derived probe or Service ports.

#### Scenario: allow omitted default listen address
- **WHEN** the input spec omits `controlPlane.defaultListenAddress`
- **THEN** validation succeeds
- **AND** the renderer may continue to normalize the value to `:8181`

#### Scenario: allow explicit host and port listen address
- **WHEN** the input spec sets `controlPlane.defaultListenAddress` to a valid `host:port`, `:port`, or bracketed IPv6 `host:port` value
- **THEN** validation succeeds

#### Scenario: reject malformed default listen address
- **WHEN** the input spec sets `controlPlane.defaultListenAddress` to a malformed, portless, non-numeric, or out-of-range value
- **THEN** validation fails before plan rendering
