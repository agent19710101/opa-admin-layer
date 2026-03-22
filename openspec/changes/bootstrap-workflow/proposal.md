# Proposal: bootstrap workflow-first OPA admin layer

## Why

The repository needs a repeatable delivery system before broader feature work. The workflow must ingest documentation, track control-plane state, plan changes via OpenSpec, implement small Go slices, validate them, and prepare GitHub sync/release work on a schedule.

## What changes

- bootstrap repository structure and documentation-understanding artifacts
- add runtime control/status files managed outside the repository
- add cycle scripts for preflight, dispatch, and closeout
- implement a first Go vertical slice for validating admin specs and rendering OPA-only plans
- add CI and release-preparation workflows
