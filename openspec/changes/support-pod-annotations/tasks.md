# Tasks

- [x] define the shared/topic pod annotation contract and keep the first slice limited to pod-template annotations only
- [x] add `controlPlane.podAnnotations` and `topic.podAnnotations` to the spec model and validation layer
- [x] merge topic pod annotations over shared control-plane defaults during Deployment rendering
- [x] render the effective annotation set into generated Deployment pod templates
- [x] add regression tests for inherited, merged, overridden, and invalid pod annotation cases
- [x] update README, architecture notes, understanding docs, examples, and changelog once implementation lands
