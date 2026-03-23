# Design: topic-specific Service overrides

## Context

The current renderer carries one shared Service configuration across every tenant/topic workload:

- `controlPlane.serviceType`
- `controlPlane.serviceAnnotations`

That keeps the contract small, but it also means topic boundaries stop at the Deployment layer. Any topic that needs a different Service exposure mode or controller metadata must be patched after render, even though the renderer already owns the Service manifest shape.

## Decision

Add two optional topic-level fields:

- `topic.serviceType`
- `topic.serviceAnnotations`

Rendering rules:

1. Start from the shared control-plane Service defaults.
2. If `topic.serviceType` is set, use it instead of `controlPlane.serviceType` for that topic Service.
3. Merge `topic.serviceAnnotations` over `controlPlane.serviceAnnotations` key-by-key so topics can replace one annotation without restating the full shared map.
4. Keep the scope intentionally narrow to the rendered Service metadata already owned by the plan renderer; do not add selector, port, namespace, or Pod/Deployment metadata controls in this slice.

## Consequences

Positive:

- topic boundaries now extend to Service exposure and controller metadata
- operators can keep a shared default and only describe the exception per topic
- the contract matches the existing topic-level `opaResources` override model, which reduces surprise

Tradeoffs:

- topic-level networking drift becomes easier, so later policy guardrails may be needed if fleets overuse the escape hatch
- validation must stay symmetric between shared and topic-level fields to avoid one path becoming looser than the other