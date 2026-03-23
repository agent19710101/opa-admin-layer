# Tasks

- [x] define the topic-level Service override contract and keep the first slice limited to `serviceType` plus `serviceAnnotations`
- [x] add `topic.serviceType` and `topic.serviceAnnotations` to the spec model and validation layer
- [x] merge topic-level Service annotations over shared control-plane annotations during plan rendering
- [x] let `topic.serviceType` override the shared Service type for the rendered tenant/topic Service
- [x] add regression tests for inherited, merged, overridden, and invalid topic Service metadata cases
- [x] update README, architecture notes, and understanding docs once implementation lands
