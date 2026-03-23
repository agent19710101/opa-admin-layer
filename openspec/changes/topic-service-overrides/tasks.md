# Tasks

- [ ] define the topic-level Service override contract and keep the first slice limited to `serviceType` plus `serviceAnnotations`
- [ ] add `topic.serviceType` and `topic.serviceAnnotations` to the spec model and validation layer
- [ ] merge topic-level Service annotations over shared control-plane annotations during plan rendering
- [ ] let `topic.serviceType` override the shared Service type for the rendered tenant/topic Service
- [ ] add regression tests for inherited, merged, overridden, and invalid topic Service metadata cases
- [ ] update README, architecture notes, and understanding docs once implementation lands