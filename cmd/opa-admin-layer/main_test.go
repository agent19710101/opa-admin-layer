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
		filepath.Join(outDir, "tenant-a", "billing", "deployment.yaml"),
		filepath.Join(outDir, "tenant-a", "billing", "service.yaml"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
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

func TestRunRenderWritesTopicServiceOverrides(t *testing.T) {
	specPath := filepath.Join(t.TempDir(), "spec.json")
	spec := `{
  "name": "demo",
  "controlPlane": {
    "baseServiceURL": "https://control.example.com",
    "serviceType": "LoadBalancer",
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
