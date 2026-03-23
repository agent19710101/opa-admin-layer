# Proposal: support Service internalTrafficPolicy controls

## Why

Generated Services already support shared/topic `serviceType`, `externalTrafficPolicy`, `sessionAffinity`, and `serviceAnnotations`. One practical in-cluster networking knob is still missing for operators that want node-local routing semantics between workloads without patching generated manifests afterward: Kubernetes `internalTrafficPolicy`.

## What changes

- add optional shared `controlPlane.internalTrafficPolicy` to the admin spec
- add optional topic-level `internalTrafficPolicy` override that inherits from the shared default when omitted
- validate `internalTrafficPolicy` against Kubernetes-supported `Cluster` and `Local` values
- render effective `internalTrafficPolicy` into generated Service manifests and cover CLI/REST/documentation paths
