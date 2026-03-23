## ADDED Requirements

### Requirement: rendered Services support inherited session affinity
The admin layer SHALL allow operators to declare Kubernetes Service `sessionAffinity` in the admin spec so generated Service manifests can express sticky-client routing without downstream patching.

#### Scenario: shared session affinity default is inherited
- **WHEN** the input spec sets `controlPlane.sessionAffinity` to `ClientIP`
- **AND** a tenant/topic omits `topic.sessionAffinity`
- **THEN** the rendered Service manifest includes `sessionAffinity: ClientIP`

#### Scenario: topic session affinity overrides the shared default
- **WHEN** the input spec sets `controlPlane.sessionAffinity` to `ClientIP`
- **AND** a tenant/topic sets `topic.sessionAffinity` to `None`
- **THEN** that topic's rendered Service manifest includes `sessionAffinity: None`
- **AND** the shared default is not rendered for that topic

#### Scenario: invalid session affinity is rejected during validation
- **WHEN** the input spec sets `controlPlane.sessionAffinity` or `topic.sessionAffinity` to a value other than `None` or `ClientIP`
- **THEN** validation fails before plan rendering with an error that points at the invalid field
