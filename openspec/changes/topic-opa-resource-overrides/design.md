# Design: topic-level OPA resource overrides

## Context

The renderer already supports shared `controlPlane.opaResources` defaults, but the next operational pain point is workload skew across topics. A billing topic may need more memory than a low-traffic support topic even when both share the same control plane.

## Decision

Treat `topic.opaResources` as a field-by-field override layer on top of `controlPlane.opaResources`:

- `requests.cpu` overrides only shared `requests.cpu`
- `requests.memory` overrides only shared `requests.memory`
- `limits.cpu` overrides only shared `limits.cpu`
- `limits.memory` overrides only shared `limits.memory`
- unspecified topic fields inherit the shared value unchanged

## Rationale

This keeps the authoring model compact and avoids repeating the entire shared resource profile whenever one topic only needs a single memory or CPU adjustment.

## Follow-up

If topic overrides begin to drift widely, future slices can add policy guardrails such as allowed ranges, ratio checks, or admission-like budgeting rules.
