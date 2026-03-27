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
		SharedServiceAccounts: []SharedServiceAccountPlan{{
			Name:         "opa-shared",
			ManifestYAML: "apiVersion: v1\nkind: ServiceAccount\nmetadata:\n  name: opa-shared\n",
		}},
		Tenants: []TenantPlan{{
			Name: "tenant-a",
			Topics: []TopicPlan{{
				Name:                       "billing",
				OPAConfigYAML:              "services:\n  controlplane:\n",
				ConfigMapManifestYAML:      "apiVersion: v1\nkind: ConfigMap\n",
				ServiceAccountManifestYAML: "apiVersion: v1\nkind: ServiceAccount\n",
				DeploymentManifestYAML:     "apiVersion: apps/v1\nkind: Deployment\n",
				ServiceManifestYAML:        "apiVersion: v1\nkind: Service\n",
				HPAManifestYAML:            "apiVersion: autoscaling/v2\nkind: HorizontalPodAutoscaler\n",
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
	if !strings.Contains(string(planJSON), `"sharedServiceAccounts"`) {
		t.Fatalf("expected plan.json to include sharedServiceAccounts, got %q", string(planJSON))
	}

	sharedServiceAccount, err := os.ReadFile(filepath.Join(outDir, "shared", "serviceaccounts", "opa-shared", "serviceaccount.yaml"))
	if err != nil {
		t.Fatalf("read shared serviceaccount.yaml: %v", err)
	}
	if string(sharedServiceAccount) != plan.SharedServiceAccounts[0].ManifestYAML {
		t.Fatalf("shared serviceaccount.yaml mismatch: got %q want %q", string(sharedServiceAccount), plan.SharedServiceAccounts[0].ManifestYAML)
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

	serviceAccount, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "serviceaccount.yaml"))
	if err != nil {
		t.Fatalf("read serviceaccount.yaml: %v", err)
	}
	if string(serviceAccount) != plan.Tenants[0].Topics[0].ServiceAccountManifestYAML {
		t.Fatalf("serviceaccount.yaml mismatch: got %q want %q", string(serviceAccount), plan.Tenants[0].Topics[0].ServiceAccountManifestYAML)
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

	hpa, err := os.ReadFile(filepath.Join(outDir, "tenant-a", "billing", "hpa.yaml"))
	if err != nil {
		t.Fatalf("read hpa.yaml: %v", err)
	}
	if string(hpa) != plan.Tenants[0].Topics[0].HPAManifestYAML {
		t.Fatalf("hpa.yaml mismatch: got %q want %q", string(hpa), plan.Tenants[0].Topics[0].HPAManifestYAML)
	}
}
