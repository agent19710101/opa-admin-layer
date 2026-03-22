# Design: configure rendered Service annotations

## Scope

This slice adds one shared control-plane map, `controlPlane.serviceAnnotations`, that is copied into every generated Service manifest.

The goal is to cover the most common operator need introduced by the new configurable Service type: attaching controller-specific metadata without editing generated YAML.

## Constraints

- keep annotations shared across all tenant/topic Services for now
- validate annotation keys before render so invalid metadata fails through both CLI and REST paths, but keep values as arbitrary strings and quote them safely in YAML output
- do not expand this slice into Deployment annotations, Pod annotations, namespace controls, or per-topic overrides
- preserve the current small-renderer shape: if the field is omitted, generated output stays unchanged

## Follow-up opportunities

If operators later need different annotation sets per topic, this shared control-plane field can remain the default while a narrower override layer is added separately.
