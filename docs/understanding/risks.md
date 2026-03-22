# Risks

- A workflow-only scaffold can drift into paperwork if dispatch scripts do not keep producing real code or validation outputs.
- Topic-scoped tenancy needs careful naming and ownership boundaries to avoid accidental policy leakage.
- OPA-only is the correct starting topology, but fast-changing data may later require bundle/discovery improvements or OPAL-like capabilities.
- A hard-coded OPA image can drift from environment-specific registry, mirroring, or approval policy unless the plan renderer exposes that pin as spec input.
- A deployment manifest that references `/config/opa-config.yaml` without provisioning it is not runnable and pushes avoidable integration work onto operators.
- Deployment health probes now assume the normalized listen address exposes OPA HTTP health endpoints on the derived port; future HTTPS or auth-protected control-plane patterns may need probe configurability.
- Rendered Service type is now configurable for ClusterIP/NodePort/LoadBalancer, but controller-specific Service annotations are still missing today, so generated YAML still requires downstream patching for common load-balancer and ingress integrations.
- Even after shared Service annotation support lands, future edge, mesh, or multi-port exposure needs may still require deeper networking controls.
- Topic labels are now validated against Kubernetes syntax before render, but future metadata expansion (annotations, selectors, namespaces) will need equally explicit guardrails to avoid generating invalid manifests.
- Generated workload object names depend on spec, tenant, and topic identifiers; without explicit validation, a single punctuation mark or oversized identifier can still produce unusable Deployment, ConfigMap, or Service manifests.
- Release automation exists locally, but GitHub push/release execution still depends on remote repo creation and permissions.
