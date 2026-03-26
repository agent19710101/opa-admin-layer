# Tasks

- [x] define the shared/topic deployment annotation contract and keep the first slice limited to Deployment metadata annotations only
- [x] add `controlPlane.deploymentAnnotations` and `topic.deploymentAnnotations` to the spec model and validation layer
- [x] merge topic deployment annotations over shared control-plane defaults during Deployment rendering
- [x] render the effective annotation set into generated Deployment metadata
- [x] add regression tests for inherited, merged, overridden, and invalid deployment annotation cases
- [x] update README, architecture notes, understanding docs, examples, and changelog once implementation lands
