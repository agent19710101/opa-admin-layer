# Tasks

- [x] define the shared/topic service account contract and keep the first slice limited to Deployment pod spec rendering only
- [x] add `controlPlane.serviceAccountName` and `topic.serviceAccountName` to the spec model, decode path, and validation layer
- [x] inherit the shared value into rendered Deployments and allow topic overrides to replace it
- [x] add regression tests for decode, invalid input, inheritance, and override behavior
- [x] update README, architecture notes, understanding docs, examples, changelog, and runtime status once implementation lands
