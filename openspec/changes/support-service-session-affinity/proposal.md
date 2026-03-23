# Proposal: support Service session affinity controls

Generated Services already support shared/topic `serviceType`, `externalTrafficPolicy`, and `serviceAnnotations`. One practical networking knob is still missing for operators that need sticky client routing across repeated requests: Kubernetes `sessionAffinity`.

Without this field in the admin spec, teams still have to patch rendered Service manifests after generation when a topic needs `ClientIP` affinity or when a shared platform default should force `None` explicitly.

## Proposed change

- add optional shared `controlPlane.sessionAffinity` to the admin spec
- add optional topic-level `sessionAffinity` override that inherits from the shared default when omitted
- validate values against the first narrow Kubernetes contract: `None` or `ClientIP`
- render effective `sessionAffinity` into generated Service manifests for CLI and REST plan output
- update docs and checked-in example specs to show the new Service metadata inheritance path
