package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRenderWithOutDirWritesPlanTree(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "configMapAnnotations": {
      "reloader.stakater.com/match": "true"
    },
    "serviceAccountName": "opa-shared",
    "serviceAccountAnnotations": {
      "eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/shared-opa"
    },
    "serviceAccountLabels": {
      "example.com/service-account-scope": "shared"
    },
    "opaResources": {
      "requests": {
        "cpu": "100m",
        "memory": "128Mi"
      }
    }
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing",
          "opaResources": {
            "requests": {
              "memory": "256Mi"
            }
          }
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	stdoutPath := filepath.Join(t.TempDir(), "plan.stdout.json")
	outDir := filepath.Join(t.TempDir(), "tree")
	if err := run([]string{"render", "-input", specPath, "-output", stdoutPath, "-outdir", outDir}); err != nil {
		t.Fatalf("run render: %v", err)
	}

	for _, path := range []string{
		stdoutPath,
		filepath.Join(outDir, "plan.json"),
		filepath.Join(outDir, "tenant-a", "billing", "opa-config.yaml"),
		filepath.Join(outDir, "tenant-a", "billing", "configmap.yaml"),
		filepath.Join(outDir, "tenant-a", "billing", "serviceaccount.yaml"),
		filepath.Join(outDir, "tenant-a", "billing", "deployment.yaml"),
		filepath.Join(outDir, "tenant-a", "billing", "service.yaml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}

	configMapBytes, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "configmap.yaml"))
	if err != nil {
		t.Fatalf("read configmap: %v", err)
	}
	configMap := string(configMapBytes)
	if !strings.Contains(configMap, `reloader.stakater.com/match: "true"`) {
		t.Fatalf("expected shared configMap annotation in rendered configmap, got %s", configMap)
	}

	serviceAccountBytes, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "serviceaccount.yaml"))
	if err != nil {
		t.Fatalf("read serviceaccount: %v", err)
	}
	serviceAccount := string(serviceAccountBytes)
	for _, expected := range []string{
		"kind: ServiceAccount",
		"name: opa-shared",
		`eks.amazonaws.com/role-arn: "arn:aws:iam::123456789012:role/shared-opa"`,
		`example.com/service-account-scope: "shared"`,
	} {
		if !strings.Contains(serviceAccount, expected) {
			t.Fatalf("expected rendered service account manifest to contain %q, got %s", expected, serviceAccount)
		}
	}

	deploymentBytes, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "deployment.yaml"))
	if err != nil {
		t.Fatalf("read deployment: %v", err)
	}
	deployment := string(deploymentBytes)
	if !strings.Contains(deployment, `cpu: "100m"`) {
		t.Fatalf("expected shared CPU request in rendered deployment, got %s", deployment)
	}
	if !strings.Contains(deployment, `memory: "256Mi"`) {
		t.Fatalf("expected topic memory override in rendered deployment, got %s", deployment)
	}
	if strings.Contains(deployment, `memory: "128Mi"`) {
		t.Fatalf("expected topic memory override to replace shared value, got %s", deployment)
	}
}

func TestRunValidateRejectsUnknownFields(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com"
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing",
          "unexpected": true
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"validate", "-input", specPath})
	if err == nil {
		t.Fatal("expected validate to fail for unknown fields")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestRunValidateAcceptsYAMLInput(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.yaml")
	spec := `name: demo
controlPlane:
  baseServiceURL: https://control.example.com
tenants:
  - name: tenant-a
    topics:
      - name: billing
`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	if err := run([]string{"validate", "-input", specPath}); err != nil {
		t.Fatalf("expected YAML validate to pass, got %v", err)
	}
}

func TestRunValidateRejectsNegativeReplicas(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "replicas": -1
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing",
          "replicas": -2
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"validate", "-input", specPath})
	if err == nil {
		t.Fatal("expected validate to fail for negative replicas")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failure, got %v", err)
	}
}

func TestRunValidateRejectsInvalidDefaultListenAddress(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "defaultListenAddress": "localhost"
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"validate", "-input", specPath})
	if err == nil {
		t.Fatal("expected validate to fail for invalid default listen address")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failure, got %v", err)
	}
}

func TestRunValidateRejectsInvalidOPAResourceQuantities(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "opaResources": {
      "requests": {
        "cpu": "ten millicores"
      }
    }
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"validate", "-input", specPath})
	if err == nil {
		t.Fatal("expected validate to fail for invalid OPA resource quantities")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failure, got %v", err)
	}
}

func TestRunRenderWritesReplicaOverrides(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "replicas": 3
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing"
        },
        {
          "name": "support",
          "replicas": 5
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "tree")
	if err := run([]string{"render", "-input", specPath, "-outdir", outDir}); err != nil {
		t.Fatalf("run render: %v", err)
	}

	billingDeployment, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "deployment.yaml"))
	if err != nil {
		t.Fatalf("read billing deployment: %v", err)
	}
	if !strings.Contains(string(billingDeployment), "replicas: 3") {
		t.Fatalf("expected shared replicas in billing deployment, got %s", string(billingDeployment))
	}

	supportDeployment, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "support", "deployment.yaml"))
	if err != nil {
		t.Fatalf("read support deployment: %v", err)
	}
	if !strings.Contains(string(supportDeployment), "replicas: 5") {
		t.Fatalf("expected topic replica override in support deployment, got %s", string(supportDeployment))
	}
	if strings.Contains(string(supportDeployment), "replicas: 3") {
		t.Fatalf("expected topic replica override to replace shared replicas, got %s", string(supportDeployment))
	}
}

func TestRunRenderWritesTopicServiceOverrides(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "serviceType": "LoadBalancer",
    "externalTrafficPolicy": "Cluster",
    "internalTrafficPolicy": "Cluster",
    "sessionAffinity": "ClientIP",
    "serviceAnnotations": {
      "example.com/scope": "shared",
      "example.com/health-check-path": "/health"
    }
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing",
          "serviceType": "NodePort",
          "externalTrafficPolicy": "Local",
          "internalTrafficPolicy": "Local",
          "sessionAffinity": "None",
          "serviceAnnotations": {
            "example.com/scope": "billing",
            "example.com/exposure": "public"
          }
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "tree")
	if err := run([]string{"render", "-input", specPath, "-outdir", outDir}); err != nil {
		t.Fatalf("run render: %v", err)
	}

	serviceBytes, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "service.yaml"))
	if err != nil {
		t.Fatalf("read service: %v", err)
	}
	service := string(serviceBytes)
	for _, expected := range []string{
		"type: NodePort",
		"externalTrafficPolicy: Local",
		"sessionAffinity: None",
		`example.com/scope: "billing"`,
		`example.com/exposure: "public"`,
		`example.com/health-check-path: "/health"`,
	} {
		if !strings.Contains(service, expected) {
			t.Fatalf("expected rendered service to contain %q, got %s", expected, service)
		}
	}
	if strings.Contains(service, "type: LoadBalancer") {
		t.Fatalf("expected topic service type override to replace shared type, got %s", service)
	}
}

func TestRunValidateRejectsOPAResourceRequestsAboveLimits(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "opaResources": {
      "requests": {
        "cpu": "1000m"
      },
      "limits": {
        "cpu": "500m"
      }
    }
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"validate", "-input", specPath})
	if err == nil {
		t.Fatal("expected validate to fail when OPA resource requests exceed limits")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected resource budget validation failure, got %v", err)
	}
}

func TestRunValidateRejectsInvalidInternalTrafficPolicyAndSessionAffinity(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "internalTrafficPolicy": "Edge",
    "sessionAffinity": "Sticky"
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"validate", "-input", specPath})
	if err == nil {
		t.Fatal("expected validate to fail for invalid sessionAffinity")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failure, got %v", err)
	}
}

func TestRunValidateRejectsExternalTrafficPolicyWithoutExternallyExposedService(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "externalTrafficPolicy": "Local"
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	err := run([]string{"validate", "-input", specPath})
	if err == nil {
		t.Fatal("expected validate to fail for incompatible externalTrafficPolicy")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Fatalf("expected validation failure, got %v", err)
	}
}

func TestRunRenderWritesTopicDeploymentAndPodAnnotations(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "deploymentAnnotations": {
      "example.com/owner": "platform",
      "example.com/revision-window": "shared"
    },
    "podAnnotations": {
      "sidecar.istio.io/inject": "false",
      "example.com/trace-sampling": "shared"
    }
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing",
          "deploymentAnnotations": {
            "example.com/revision-window": "billing",
            "example.com/rollout": "canary"
          },
          "podAnnotations": {
            "example.com/trace-sampling": "billing",
            "example.com/debug": "enabled"
          }
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "tree")
	if err := run([]string{"render", "-input", specPath, "-outdir", outDir}); err != nil {
		t.Fatalf("run render: %v", err)
	}

	deploymentBytes, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "deployment.yaml"))
	if err != nil {
		t.Fatalf("read deployment: %v", err)
	}
	deployment := string(deploymentBytes)
	for _, expected := range []string{
		`example.com/owner: "platform"`,
		`example.com/revision-window: "billing"`,
		`example.com/rollout: "canary"`,
		`sidecar.istio.io/inject: "false"`,
		`example.com/trace-sampling: "billing"`,
		`example.com/debug: "enabled"`,
	} {
		if !strings.Contains(deployment, expected) {
			t.Fatalf("expected rendered deployment to contain %q, got %s", expected, deployment)
		}
	}
	if strings.Contains(deployment, `example.com/revision-window: "shared"`) {
		t.Fatalf("expected topic deployment annotation override to replace shared value, got %s", deployment)
	}
	if strings.Contains(deployment, `example.com/trace-sampling: "shared"`) {
		t.Fatalf("expected topic pod annotation override to replace shared value, got %s", deployment)
	}
}

func TestRunRenderWritesAutoscalingManifest(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "opaResources": {
      "requests": {
        "cpu": "100m"
      }
    },
    "autoscaling": {
      "minReplicas": 2,
      "maxReplicas": 5,
      "targetCPUUtilizationPercentage": 75
    }
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "tree")
	if err := run([]string{"render", "-input", specPath, "-outdir", outDir}); err != nil {
		t.Fatalf("run render: %v", err)
	}

	hpa, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "hpa.yaml"))
	if err != nil {
		t.Fatalf("read hpa: %v", err)
	}
	if !strings.Contains(string(hpa), "kind: HorizontalPodAutoscaler") {
		t.Fatalf("expected rendered hpa manifest, got %s", string(hpa))
	}
}

func TestRunRenderWritesSharedServiceAccountArtifactForSharedBindings(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "serviceAccountName": "opa-shared",
    "serviceAccountAnnotations": {
      "example.com/source": "shared"
    },
    "serviceAccountLabels": {
      "example.com/scope": "shared"
    }
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing"
        },
        {
          "name": "support"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "tree")
	if err := run([]string{"render", "-input", specPath, "-outdir", outDir}); err != nil {
		t.Fatalf("run render: %v", err)
	}

	sharedServiceAccountPath := filepath.Join(outDir, "shared", "serviceaccounts", "opa-shared", "serviceaccount.yaml")
	sharedServiceAccountBytes, err := os.ReadFile(sharedServiceAccountPath)
	if err != nil {
		t.Fatalf("read shared serviceaccount: %v", err)
	}
	for _, expected := range []string{"name: opa-shared", `example.com/source: "shared"`, `example.com/scope: "shared"`} {
		if !strings.Contains(string(sharedServiceAccountBytes), expected) {
			t.Fatalf("expected shared serviceaccount.yaml to contain %q, got %s", expected, string(sharedServiceAccountBytes))
		}
	}

	for _, topic := range []string{"billing", "support"} {
		if _, err := os.Stat(filepath.Join(outDir, "tenant-a", topic, "serviceaccount.yaml")); !os.IsNotExist(err) {
			t.Fatalf("expected shared-binding topic %s to omit topic-scoped serviceaccount.yaml, got err=%v", topic, err)
		}
		deploymentBytes, err := os.ReadFile(filepath.Join(outDir, "tenant-a", topic, "deployment.yaml"))
		if err != nil {
			t.Fatalf("read deployment for %s: %v", topic, err)
		}
		if !strings.Contains(string(deploymentBytes), "serviceAccountName: opa-shared") {
			t.Fatalf("expected deployment for %s to keep serviceAccountName binding, got %s", topic, string(deploymentBytes))
		}
	}
}

func TestRunRenderWritesTopicListenAddressOverride(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "defaultListenAddress": ":8181"
  },
  "tenants": [
    {
      "name": "tenant-a",
      "topics": [
        {
          "name": "billing",
          "listenAddress": "127.0.0.1:8282"
        }
      ]
    }
  ]
}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	outDir := filepath.Join(t.TempDir(), "tree")
	if err := run([]string{"render", "-input", specPath, "-outdir", outDir}); err != nil {
		t.Fatalf("run render: %v", err)
	}

	deploymentBytes, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "deployment.yaml"))
	if err != nil {
		t.Fatalf("read deployment: %v", err)
	}
	deployment := string(deploymentBytes)
	if !strings.Contains(deployment, "containerPort: 8282") || !strings.Contains(deployment, "--addr=127.0.0.1:8282") {
		t.Fatalf("expected deployment to use topic listenAddress override, got %s", deployment)
	}

	serviceBytes, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "service.yaml"))
	if err != nil {
		t.Fatalf("read service: %v", err)
	}
	service := string(serviceBytes)
	if !strings.Contains(service, "port: 8282") || !strings.Contains(service, "targetPort: 8282") {
		t.Fatalf("expected service to use topic listenAddress override, got %s", service)
	}
}
