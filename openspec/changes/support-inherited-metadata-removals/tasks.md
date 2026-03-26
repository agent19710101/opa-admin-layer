## 1. Spec and validation

- [x] 1.1 Add topic-level removal-list fields for inherited Service, ConfigMap, Deployment, and pod annotation/label metadata.
- [x] 1.2 Validate removal-list entries as Kubernetes metadata keys.

## 2. Rendering and docs

- [x] 2.1 Apply removal lists after shared/topic metadata merge so inherited keys can be cleared.
- [x] 2.2 Keep built-in rendered identity labels immutable and present after removals.
- [x] 2.3 Update docs, architecture notes, and examples/tests to reflect deletion semantics.

## 3. Verification

- [x] 3.1 Add or extend unit tests for decode, render, removal, and invalid-key cases.
- [x] 3.2 Run focused Go test coverage for touched packages and CLI/example validation.
