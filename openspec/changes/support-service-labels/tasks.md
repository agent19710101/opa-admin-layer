## 1. Spec and validation

- [x] 1.1 Add shared and topic-level `serviceLabels` fields to the admin spec model.
- [x] 1.2 Validate Service label keys and values for both shared and topic scopes.

## 2. Rendering and examples

- [x] 2.1 Render effective Service labels onto `Service.metadata.labels` only.
- [x] 2.2 Keep built-in identity labels immutable when custom Service labels overlap.
- [x] 2.3 Update checked-in JSON/YAML examples and README snippets.

## 3. Verification

- [x] 3.1 Add or extend unit tests for merge, scoping, immutability, and invalid label cases.
- [x] 3.2 Run Go test coverage for the touched packages and CLI example validation.
