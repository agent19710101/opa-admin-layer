# Proposal: support Service externalTrafficPolicy controls

## Why

The renderer already owns Kubernetes Service type and annotation metadata, and operators can now override those defaults per topic. One practical gap remains: externally exposed Services often need `externalTrafficPolicy: Local` or `Cluster` to control source IP preservation and node-local routing behavior. Without that knob, teams still have to patch rendered Service manifests after generation.

## Change

- add optional shared `controlPlane.externalTrafficPolicy` to the admin spec
- add optional topic-level `externalTrafficPolicy` override that inherits from the shared default when omitted
- validate allowed values (`Cluster` or `Local`) in shared and topic paths
- validate the effective Service combination so `externalTrafficPolicy` is only accepted when the final Service type is `NodePort` or `LoadBalancer`
- render `externalTrafficPolicy` into generated Service manifests and cover CLI/REST/documentation paths

## Impact

Operators can express one of the most common Service traffic-policy controls directly in the admin contract without widening scope into ports, selectors, or broader ingress abstractions.
