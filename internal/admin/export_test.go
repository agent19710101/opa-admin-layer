package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWritePlanTree(t *testing.T) {
	plan := Plan{
		GeneratedAt: "2026-03-22T00:00:00Z",
		Name:        "demo",
		Topology:    "opa-only",
		Tenants: []TenantPlan{{
			Name: "tenant-a",
			Topics: []TopicPlan{{
				Name:                   "billing",
				OPAConfigYAML:          "services:\n  controlplane:\n",
				ConfigMapManifestYAML:  "apiVersion: v1\nkind: ConfigMap\n",
				DeploymentManifestYAML: "apiVersion: apps/v1\nkind: Deployment\n",
			}},
		}},
	}

	outDir := t.TempDir()
	if err := WritePlanTree(plan, outDir); err != nil {
		t.Fatalf("WritePlanTree returned error: %v", err)
	}

	planJSON, err := os.ReadFile(filepath.Join(outDir, "plan.json"))
	if err != nil {
		t.Fatalf("read plan.json: %v", err)
	}
	if !strings.Contains(string(planJSON), `"name": "demo"`) {
		t.Fatalf("expected plan.json to include plan name, got %q", string(planJSON))
	}

	opaConfig, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "opa-config.yaml"))
	if err != nil {
		t.Fatalf("read opa-config.yaml: %v", err)
	}
	if string(opaConfig) != plan.Tenants[0].Topics[0].OPAConfigYAML {
		t.Fatalf("opa-config.yaml mismatch: got %q want %q", string(opaConfig), plan.Tenants[0].Topics[0].OPAConfigYAML)
	}

	configMap, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "configmap.yaml"))
	if err != nil {
		t.Fatalf("read configmap.yaml: %v", err)
	}
	if string(configMap) != plan.Tenants[0].Topics[0].ConfigMapManifestYAML {
		t.Fatalf("configmap.yaml mismatch: got %q want %q", string(configMap), plan.Tenants[0].Topics[0].ConfigMapManifestYAML)
	}

	deployment, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "deployment.yaml"))
	if err != nil {
		t.Fatalf("read deployment.yaml: %v", err)
	}
	if string(deployment) != plan.Tenants[0].Topics[0].DeploymentManifestYAML {
		t.Fatalf("deployment.yaml mismatch: got %q want %q", string(deployment), plan.Tenants[0].Topics[0].DeploymentManifestYAML)
	}

	service, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "service.yaml"))
	if err != nil {
		t.Fatalf("read service.yaml: %v", err)
	}
	if string(service) != plan.Tenants[0].Topics[0].ServiceManifestYAML {
		t.Fatalf("service.yaml mismatch: got %q want %q", string(service), plan.Tenants[0].Topics[0].ServiceManifestYAML)
	}
}
