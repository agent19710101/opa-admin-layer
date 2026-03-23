## ADDED Requirements

### Requirement: Shared Service internal traffic policy

The admin layer SHALL allow operators to declare Kubernetes Service `internalTrafficPolicy` in the admin spec so generated Service manifests can express in-cluster routing behavior without downstream patching.

#### Scenario: shared internal traffic policy is inherited

- **WHEN** the input spec sets `controlPlane.internalTrafficPolicy` to `Local`
- **AND** a tenant/topic omits `topic.internalTrafficPolicy`
- **THEN** the rendered Service manifest includes `internalTrafficPolicy: Local`

#### Scenario: topic internal traffic policy overrides the shared default

- **WHEN** the input spec sets `controlPlane.internalTrafficPolicy` to `Cluster`
- **AND** a tenant/topic sets `topic.internalTrafficPolicy` to `Local`
- **THEN** that topic's rendered Service manifest includes `internalTrafficPolicy: Local`

#### Scenario: invalid internal traffic policy is rejected

- **WHEN** the input spec sets `controlPlane.internalTrafficPolicy` or `topic.internalTrafficPolicy` to a value other than `Cluster` or `Local`
- **THEN** validation fails before render
