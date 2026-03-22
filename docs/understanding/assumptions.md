# Assumptions

- The first delivery slice can focus on plan generation instead of full policy publishing.
- JSON is sufficient for the initial admin spec format because it works cleanly for CLI and REST ingestion.
- Kubernetes deployment YAML is a useful first deployment target because the research pack emphasizes sidecar patterns.
- GitHub repo creation and push can proceed automatically because local git/GitHub actions do not require confirmation in this workspace.
- OpenSpec CLI is not required if the repository mirrors the expected structure manually.
