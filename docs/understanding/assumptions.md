# Assumptions

- The first delivery slice can focus on plan generation instead of full policy publishing.
- Admin spec ingestion should support both JSON and YAML because operators are likely to author deployment-facing configuration in YAML even when the renderer also serves strict JSON clients.
- Kubernetes deployment YAML is a useful first deployment target because the research pack emphasizes sidecar patterns.
- GitHub repo creation and push can proceed automatically because local git/GitHub actions do not require confirmation in this workspace.
- OpenSpec CLI is not required if the repository mirrors the expected structure manually.
- The smallest useful next autoscaling slice after CPU/memory metrics is stabilization-window control, because it covers a common operator need without opening arbitrary HPA policy passthrough.
