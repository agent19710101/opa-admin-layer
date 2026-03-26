# Tasks

- [x] define the shared/topic pod label contract and keep the first slice limited to pod-template labels only
- [x] add `controlPlane.podLabels` and `topic.podLabels` to the spec model and validation layer
- [x] merge topic pod labels over shared control-plane defaults during Deployment rendering
- [x] render the effective label set into generated Deployment pod templates without allowing built-in selector labels to be overridden
- [x] add regression tests for inherited, merged, overridden, and invalid pod label cases
- [x] update README, architecture notes, understanding docs, examples, and changelog once implementation lands
