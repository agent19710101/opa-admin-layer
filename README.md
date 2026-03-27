# opa-admin-layer

OPA administration layer with a workflow-first delivery model.

## Product direction

Locked decisions:

- topology: OPA-only
- implementation language: Go-only for product code
- admin surface: CLI + REST API
- tenant model: multi-tenant topic-scoped

The workflow is the primary product. Application features are shipped as outputs of the workflow.

## Repository layout

- `cmd/opa-admin-layer`: CLI entrypoint
- `internal/admin`: tenant/topic scoped spec validation and plan rendering
- `internal/httpapi`: REST API for validation and plan generation
- `docs/understanding`: documentation-ingestion artifacts
- `deploy/examples`: runnable example admin specs
- `openspec`: change planning
- `scripts/cycle`: workflow scripts
- `scripts/automation`: thin phase runner wrapper
- `.github/workflows`: CI and release preparation

## Runtime state

Runtime state is intentionally stored outside the repository.

Defaults:

- state directory: `${XDG_STATE_HOME:-$HOME/.local/state}/opa-admin-layer`
- override: set `OPA_ADMIN_LAYER_STATE_DIR`

The runtime state directory holds the active project, project queue, switch request, cycle status, summary, blockers, and logs.

## Quick start

```bash
make test
./bin/opa-admin-layer validate -input deploy/examples/dev-spec.json
./bin/opa-admin-layer render -input deploy/examples/dev-spec.json
./bin/opa-admin-layer render -input deploy/examples/dev-spec.json -outdir ./tmp/plan
# YAML is also supported for CLI input.
./bin/opa-admin-layer validate -input deploy/examples/dev-spec.yaml
./bin/opa-admin-layer serve -addr :8080
```

Build first:

```bash
make build
```

The repository now ships matching example specs in both formats:

- `deploy/examples/dev-spec.json`
- `deploy/examples/dev-spec.yaml`

Example API usage:

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/v1/validate \
  -H 'content-type: application/json' \
  --data @deploy/examples/dev-spec.json
curl -s http://localhost:8080/v1/validate \
  -H 'content-type: application/yaml' \
  --data-binary @deploy/examples/dev-spec.yaml
curl -s http://localhost:8080/v1/plans \
  -H 'content-type: application/json' \
  --data @deploy/examples/dev-spec.json
```

Accepted REST request media types for `/v1/validate` and `/v1/plans`:

- `application/json`
- `application/yaml`
- `application/x-yaml`
- `text/yaml`
- `text/x-yaml`
- empty `Content-Type` is also accepted for simple callers and falls back to payload sniffing

## Workflow loop

A scheduler should trigger exactly three jobs during the active window:

- `dev-preflight` at `55 19 * * *` Europe/Warsaw
- `dev-cycle-dispatch` at `*/15 20-23,0-4 * * *` Europe/Warsaw
- `dev-closeout` at `0 5 * * *` Europe/Warsaw

Use `scripts/automation/run-phase.sh` as the repo entrypoint.

The dispatcher advances one phase per run in this order:

1. preflight and control-plane check
2. documentation ingestion
3. architecture/repo layout
4. OpenSpec update
5. smallest useful implementation slice
6. verification
7. sync and release preparation
8. persist cycle state

## Current vertical slice

The first shipped slice validates a tenant/topic scoped admin spec and renders an OPA-only plan containing:

- strict JSON or YAML decoding for CLI and REST input (unknown fields are rejected early in both formats)
- configurable but pinned OPA image selection via `controlPlane.opaImage`
- optional shared `controlPlane.imagePullPolicy` plus topic-level overrides so rendered OPA Deployments can express Kubernetes image pull behavior without downstream patches
- optional shared `controlPlane.autoscaling` plus topic-level overrides so generated workloads can emit Kubernetes HorizontalPodAutoscaler manifests with CPU and/or memory utilization targets and optional scale-up/scale-down stabilization windows, `selectPolicy`, and explicit scaling `policies` without downstream patching, with effective matching `opaResources.requests.cpu` and/or `opaResources.requests.memory` required for configured autoscaling metrics
- normalized tenant/topic inventory
- per-topic OPA bundle URL
- generated OPA config YAML
- generated Kubernetes ConfigMap YAML carrying the OPA config
- generated Kubernetes deployment YAML for a sidecar-style OPA deployment
- declared OPA container ports and default readiness/liveness probes derived from the normalized listen address
- propagated per-topic Kubernetes labels from the admin spec into generated manifests, with Kubernetes label syntax validation at ingest time
- rendered Kubernetes Deployment/ConfigMap/Service names validated up front so spec, tenant, and topic identifiers cannot produce invalid workload object names
- optional shared rendered Kubernetes namespace via `controlPlane.namespace` so generated ConfigMap, Deployment, and Service manifests can land outside the default namespace without downstream patching
- configurable rendered Kubernetes Service type via `controlPlane.serviceType`, defaulting to `ClusterIP` and rejecting unsupported values early
- optional shared rendered Service annotations via `controlPlane.serviceAnnotations` for controller/load-balancer integration metadata without post-render patching
- optional shared `controlPlane.serviceLabels` plus topic-level overrides so rendered Services can carry object-scoped labels without mutating Deployments, ConfigMaps, or pod templates
- optional topic-level `removeServiceAnnotations`, `removeServiceLabels`, `removeConfigMapAnnotations`, `removeConfigMapLabels`, `removeDeploymentAnnotations`, `removeDeploymentLabels`, `removePodAnnotations`, and `removePodLabels` lists so inherited metadata defaults can be cleared back to absent state without downstream patching
- optional shared `controlPlane.configMapAnnotations` plus topic-level `configMapAnnotations` overrides so rendered ConfigMaps can carry reloader, ownership, or GitOps metadata without downstream patching
- optional shared `controlPlane.configMapLabels` plus topic-level overrides so rendered ConfigMaps can carry object-scoped labels without mutating Services, Deployments, or pod templates
- optional shared `controlPlane.deploymentAnnotations` plus topic-level overrides so rendered Deployments can carry rollout, ownership, or GitOps metadata without downstream patching
- optional shared `controlPlane.deploymentLabels` plus topic-level overrides so rendered Deployment metadata can carry rollout tracking, ownership, or GitOps labels without mutating Services, ConfigMaps, or pod templates
- optional shared `controlPlane.podAnnotations` plus topic-level overrides so rendered OPA pod templates can carry mesh, tracing, or sidecar-injection metadata without downstream patching
- optional shared `controlPlane.podLabels` plus topic-level overrides so rendered OPA pod templates can carry pod-only discovery, policy, or workload-class labels without mutating Services or ConfigMaps
- optional shared `controlPlane.serviceAccountName` plus topic-level overrides so rendered OPA Deployments can bind to explicit Kubernetes workload identities without downstream patches
- rendered `ServiceAccount` YAML whenever a topic resolves a non-empty effective `serviceAccountName`, keeping exported workload bundles self-contained for the common service-account provisioning path
- validation now rejects repeated effective `serviceAccountName` values across topics because rendered `ServiceAccount` ownership is single-topic only today
- optional shared `controlPlane.serviceAccountAnnotations` plus topic-level overrides and `removeServiceAccountAnnotations` so rendered `ServiceAccount` objects can carry IAM/workload-identity metadata without downstream patches
- optional shared `controlPlane.serviceAccountLabels` plus topic-level overrides and `removeServiceAccountLabels` so rendered `ServiceAccount` objects can carry ownership, GitOps, and policy labels without downstream patches
- optional shared `controlPlane.imagePullPolicy` plus topic-level overrides so rendered OPA Deployments can express `Always`, `IfNotPresent`, or `Never` image pull behavior without downstream patches
- optional shared `controlPlane.automountServiceAccountToken` plus topic-level overrides so rendered OPA Deployments can explicitly keep or disable service-account token projection without downstream patches
- optional shared `controlPlane.externalTrafficPolicy` plus topic-level overrides so externally exposed Services can preserve source-aware routing behavior without downstream patching
- optional shared `controlPlane.internalTrafficPolicy` plus topic-level overrides so generated Services can steer in-cluster node-local routing (`Cluster` or `Local`) without downstream patching
- optional shared `controlPlane.sessionAffinity` plus topic-level overrides so generated Services can express sticky-client routing (`None` or `ClientIP`) without downstream patching
- optional shared OPA container CPU/memory requests and limits via `controlPlane.opaResources` so generated Deployments can carry baseline scheduling defaults
- optional per-topic `opaResources` overrides that merge over shared defaults, letting one topic raise/lower CPU or memory without restating the full resource profile
- Kubernetes quantity syntax validation for both shared and per-topic `opaResources` so malformed CPU/memory values fail early in CLI and REST validation paths
- effective OPA resource budget validation so shared and inherited topic CPU/memory requests cannot exceed their matching limits after merge
- absolute HTTP(S) validation for `controlPlane.baseServiceURL` so rendered bundle URLs and OPA config always point at a real control-plane endpoint shape
- explicit `controlPlane.defaultListenAddress` validation for `:port`, `host:port`, and bracketed IPv6 `host:port` so rendered `--addr`, Deployment probe ports, and Service ports cannot drift apart on malformed input

When `render` is called with `-outdir`, it also materializes:

- `plan.json` at the output root
- `<tenant>/<topic>/opa-config.yaml`
- `<tenant>/<topic>/configmap.yaml`
- `<tenant>/<topic>/serviceaccount.yaml` when an effective `serviceAccountName` is configured
- `<tenant>/<topic>/deployment.yaml`
- `<tenant>/<topic>/service.yaml`
- `<tenant>/<topic>/hpa.yaml` when autoscaling is configured

This slice is exposed through both the CLI and the REST API.

Example shared namespace, shared/topic ConfigMap metadata, Service metadata, inherited/overridden ServiceAccount annotations, labels, and token automount, inherited/overridden external and internal traffic policy, per-topic Service overrides, shared OPA resource defaults, and a per-topic resource override (using standard Kubernetes quantity strings):

```json
{
  "controlPlane": {
    "replicas": 2,
    "serviceType": "LoadBalancer",
    "externalTrafficPolicy": "Cluster",
    "internalTrafficPolicy": "Cluster",
    "sessionAffinity": "ClientIP",
    "serviceAnnotations": {
      "service.beta.kubernetes.io/aws-load-balancer-scheme": "internal",
      "example.com/health-check-path": "/health?plugins"
    },
    "configMapAnnotations": {
      "reloader.stakater.com/match": "true",
      "example.com/source": "generated"
    },
    "configMapLabels": {
      "example.com/config-scope": "shared",
      "example.com/team": "platform"
    },
    "deploymentAnnotations": {
      "example.com/owner": "platform",
      "example.com/revision-window": "shared"
    },
    "deploymentLabels": {
      "example.com/release-track": "shared",
      "example.com/team": "platform"
    },
    "podAnnotations": {
      "sidecar.istio.io/inject": "false",
      "example.com/trace-sampling": "shared"
    },
    "podLabels": {
      "example.com/workload-class": "shared",
      "example.com/team": "platform"
    },
    "serviceAccountName": "opa-shared",
    "serviceAccountAnnotations": {
      "eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/shared-opa"
    },
    "serviceAccountLabels": {
      "example.com/service-account-scope": "shared",
      "example.com/team": "platform"
    },
    "imagePullPolicy": "IfNotPresent",
    "automountServiceAccountToken": false,
    "opaResources": {
      "requests": {
        "cpu": "100m",
        "memory": "128Mi"
      },
      "limits": {
        "memory": "512Mi"
      }
    }
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing",
          "replicas": 4,
          "serviceType": "NodePort",
          "externalTrafficPolicy": "Local",
          "internalTrafficPolicy": "Local",
          "sessionAffinity": "None",
          "serviceAnnotations": {
            "example.com/health-check-path": "/billing-health",
            "example.com/exposure": "public"
          },
          "configMapAnnotations": {
            "example.com/source": "billing",
            "example.com/team": "payments"
          },
          "configMapLabels": {
            "example.com/config-scope": "billing",
            "example.com/ring": "canary"
          },
          "deploymentAnnotations": {
            "example.com/revision-window": "billing",
            "example.com/rollout": "canary"
          },
          "deploymentLabels": {
            "example.com/release-track": "billing",
            "example.com/ring": "canary"
          },
          "podAnnotations": {
            "example.com/trace-sampling": "billing",
            "example.com/debug": "enabled"
          },
          "podLabels": {
            "example.com/workload-class": "billing",
            "example.com/team": "payments"
          },
          "serviceAccountName": "billing-opa",
          "serviceAccountAnnotations": {
            "example.com/source": "billing",
            "example.com/team": "payments"
          },
          "removeServiceAccountAnnotations": [
            "eks.amazonaws.com/role-arn"
          ],
          "serviceAccountLabels": {
            "example.com/service-account-scope": "billing",
            "example.com/ring": "canary"
          },
          "removeServiceAccountLabels": [
            "example.com/team"
          ],
          "imagePullPolicy": "Always",
          "automountServiceAccountToken": true,
          "opaResources": {
            "requests": {
              "memory": "256Mi"
            }
          }
        }
      ]
    }
  ]
}
```

Topic metadata can also explicitly clear inherited object-scoped keys with removal lists such as `removeServiceLabels`, `removeConfigMapAnnotations`, `removeServiceAccountAnnotations`, `removeServiceAccountLabels`, or `removePodLabels` when one workload needs a shared default to end absent.

Autoscaling can use CPU utilization targets, memory utilization targets, or both. Behavior tuning can also set per-direction stabilization windows, `selectPolicy`, and explicit scaling `policies` with `Pods`/`Percent` step definitions. Any autoscaled workload must have effective inherited `opaResources.requests.cpu` and/or `opaResources.requests.memory` values for the metrics it configures.
