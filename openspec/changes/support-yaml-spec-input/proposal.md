# Proposal: support YAML admin spec input

## Why

The admin layer currently only accepts JSON specs even though its primary output is Kubernetes YAML artifacts and operators commonly manage deployment inputs in YAML. That mismatch forces users to translate or maintain parallel spec formats before they can use the CLI or REST API.

## Change

- accept YAML admin specs in the shared decode path used by CLI and REST flows
- preserve strict field validation so unknown spec fields still fail early for both JSON and YAML
- add regression coverage for CLI, shared admin decoding, and REST validation/plan generation paths
- check in a repository-owned YAML example spec that mirrors the existing JSON example so operators can exercise the supported path directly
- update repository docs and understanding notes to treat YAML as a supported ingestion format instead of an open question

## Impact

Operators can submit the same admin intent in either JSON or YAML without post-processing, while the validation contract stays strict enough to catch misspelled fields before render.
