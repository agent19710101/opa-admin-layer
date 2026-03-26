## 1. Spec and validation

- [x] 1.1 Add shared and topic-level `imagePullPolicy` fields to the admin spec model.
- [x] 1.2 Validate `imagePullPolicy` against `Always`, `IfNotPresent`, and `Never`.

## 2. Rendering and examples

- [x] 2.1 Render effective image pull policy onto the generated OPA Deployment only.
- [x] 2.2 Inherit the shared value by default and let topic overrides replace it.
- [x] 2.3 Update checked-in JSON/YAML examples and README snippets.

## 3. Verification

- [x] 3.1 Add or extend unit tests for inheritance, omission-by-default, and invalid values.
- [x] 3.2 Run Go tests for the touched packages and validate the checked-in example specs.
