# Tasks

- [x] define the shared/topic automount contract and keep the first slice limited to Deployment pod spec rendering only
- [x] add `controlPlane.automountServiceAccountToken` and `topic.automountServiceAccountToken` to the spec model and decode path
- [x] inherit the shared value into rendered Deployments and allow topic overrides to replace it
- [x] add regression tests for decode, inheritance, and rendered manifest behavior, including explicit `false` handling
- [x] update README, architecture notes, understanding docs, examples, changelog, and runtime status once implementation lands
