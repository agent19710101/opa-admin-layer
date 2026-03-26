# Proposal: support shared ConfigMap annotations

## Why

The renderer already owns the generated ConfigMap that carries `opa-config.yaml`, but operators still cannot attach ConfigMap-object metadata for common integrations such as reload controllers, ownership markers, or GitOps bookkeeping. That leaves one more routine patch step outside the admin contract even though the repository already treats Service, Deployment, and pod metadata as first-class operator concerns.

## What changes

- add optional shared `controlPlane.configMapAnnotations` to the admin spec
- validate annotation keys with the existing Kubernetes metadata-key contract
- render the configured annotations into every generated ConfigMap manifest
- keep the first slice intentionally narrow: shared ConfigMap object annotations only, with no topic overrides or arbitrary ConfigMap customization

## Impact

Operators can attach common ConfigMap metadata without post-render patching, while the contract stays small and consistent with the existing narrow metadata controls in the admin layer.
