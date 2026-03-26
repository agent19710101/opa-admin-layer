## ADDED Requirements

### Requirement: Topic metadata removal lists
The admin spec SHALL allow a topic to declare removal lists for inherited Service, ConfigMap, Deployment, and pod annotation/label metadata using Kubernetes-valid metadata keys.

#### Scenario: Removal-list entries validate
- **WHEN** a topic sets any `remove*Annotations` or `remove*Labels` field
- **THEN** validation accepts Kubernetes-valid metadata keys
- **AND** validation rejects invalid or empty keys

### Requirement: Removal lists clear inherited metadata
The renderer SHALL apply topic removal lists after shared/topic metadata merge so inherited keys can end absent for one topic.

#### Scenario: Topic clears a shared metadata key
- **GIVEN** shared control-plane metadata for a rendered Service, ConfigMap, Deployment, or pod template
- **AND** a topic removes one of those keys through the matching removal list
- **WHEN** a plan is rendered
- **THEN** the matching object metadata omits the removed key
- **AND** renderer-owned built-in identity labels remain present
