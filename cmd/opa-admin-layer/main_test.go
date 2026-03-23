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
    "baseServiceURL": "https://control.example.com"
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
