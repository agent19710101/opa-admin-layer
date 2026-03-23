# Proposal: validate control-plane base service URL

## Why

The admin contract currently only checks that `controlPlane.baseServiceURL` is non-empty. That still lets relative URLs, unsupported schemes, or malformed endpoints flow into bundle URL rendering and OPA config output. Operators then get a superficially valid plan that fails later when OPA or downstream tooling tries to use the control-plane endpoint.

## Change

- validate `controlPlane.baseServiceURL` as an absolute `http` or `https` URL before render
- reject malformed URLs, relative URLs, missing hosts, and unsupported schemes through the shared validation path used by CLI and REST flows
- add regression coverage for valid and invalid URL shapes
- update docs/design notes to describe the stricter contract

## Impact

Bundle URL and OPA config rendering now depend on a control-plane input that already matches the shape the generated artifacts expect. This keeps the admin layer’s contract closer to the real runtime contract and shifts a common operator error from apply/runtime time to validation time.
