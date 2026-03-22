# Proposal: strict spec decoding

## Why

The administration layer currently accepts unknown JSON fields when loading specs from files and when receiving specs over the REST API. That makes operator typos easy to miss and weakens validation feedback.

## Change

- reject unknown JSON fields for CLI file input
- reject unknown JSON fields for REST API input
- add regression tests for both paths

## Impact

Operators get earlier, clearer feedback when a spec includes unsupported keys. This keeps the admin surface safer without changing the rendered plan model.
