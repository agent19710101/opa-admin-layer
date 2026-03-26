## Why

Rendered Services, ConfigMaps, Deployments, and pod templates already support shared plus topic-level metadata inheritance, but operators still cannot explicitly remove a shared key for one topic. That leaves downstream patching in place for the simplest "shared by default, absent here" cases.

## What Changes

- add topic-level removal lists for inherited Service, ConfigMap, Deployment, and pod annotation/label maps
- validate removal-list entries with the existing Kubernetes metadata-key contract
- apply removals after shared+topic merge so topics can drop inherited keys back to absent state
- keep built-in rendered identity labels immutable and always present
- refresh docs and tests to show the new deletion semantics

## Impact

- removes another common manifest patch point without widening into arbitrary object customization
- keeps inheritance explicit and reviewable in the admin spec instead of hiding it in downstream overlays
- preserves selector and identity stability because built-in labels still come from the renderer, not user-controlled removals
