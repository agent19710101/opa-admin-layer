# Proposal: propagate topic labels into rendered manifests

## Why

The admin spec already allows per-topic labels, and the plan JSON preserves them, but the generated Kubernetes manifests currently ignore that metadata. Operators lose the ability to carry environment, ownership, or routing labels into the ConfigMap and Deployment artifacts without patching the output after render.

## Change

- preserve the existing built-in Kubernetes labels emitted by the renderer
- merge user-provided topic labels into generated ConfigMap and Deployment metadata
- apply the merged labels to the Deployment pod template as well
- keep renderer output deterministic and prevent topic labels from overriding built-in identity labels
- add regression tests and update example/docs to show the rendered behavior

## Impact

Rendered manifests stay self-describing and easier to integrate with cluster policy, inventory, and observability tooling while keeping the current stable identity labels intact.
