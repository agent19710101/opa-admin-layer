# Proposal: configure rendered Service annotations

## Why

The renderer now emits a Kubernetes Service for every tenant/topic OPA workload and can switch between `ClusterIP`, `NodePort`, and `LoadBalancer`. That closes the basic exposure gap, but operators still need controller-specific annotations for common real deployments such as internal load balancers, health-check tuning, or ingress/controller integration. Without first-class annotation support, generated YAML still needs downstream patching in the exact area where the renderer already owns Service shape.

## Change

- add optional `controlPlane.serviceAnnotations` to the admin spec as a string map
- validate annotation keys and values before render using the same Kubernetes metadata constraints already enforced for labels
- render configured annotations into every generated Service manifest
- keep the scope intentionally small: Service metadata only, with no Deployment or Pod annotation support in this slice
- add regression tests and docs for default, configured, and invalid annotation cases

## Impact

Generated plans stay minimal by default while covering the most common Kubernetes Service integration point that operators currently have to patch by hand.
