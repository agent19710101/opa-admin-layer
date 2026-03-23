## 1. Spec and validation

- [x] define the shared/topic `internalTrafficPolicy` contract and keep the slice limited to `Cluster` and `Local`
- [x] add shared and topic-level spec fields, normalization, and validation for `internalTrafficPolicy`

## 2. Rendering and tests

- [x] render inherited/topic-overridden `internalTrafficPolicy` into Service manifests
- [x] add or update regression coverage across admin, CLI, and HTTP validation/rendering paths

## 3. Docs and examples

- [x] update README, architecture notes, understanding docs, and example specs to document `internalTrafficPolicy` support
