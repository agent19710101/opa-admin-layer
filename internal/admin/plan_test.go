package admin

import (
	"strings"
	"testing"
)

func TestValidateRejectsDuplicateTenantAndTopic(t *testing.T) {
	spec := Specification{
		Name:         "demo",
		ControlPlane: ControlPlane{BaseServiceURL: "https://control.example.com"},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}, {Name: "billing"}},
		}, {
			Name:   "tenant-a",
			Topics: []Topic{{Name: "catalog"}},
		}},
	}

	issues := Validate(spec)
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d: %#v", len(issues), issues)
	}
}

func TestBuildPlanAppliesDefaults(t *testing.T) {
	spec := Specification{
		Name:         "demo",
		ControlPlane: ControlPlane{BaseServiceURL: "https://control.example.com"},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	if got, want := plan.Tenants[0].Topics[0].BundleURL, "https://control.example.com/bundles/tenant-a/billing.tar.gz"; got != want {
		t.Fatalf("bundle URL mismatch: got %q want %q", got, want)
	}
	if !strings.Contains(plan.Tenants[0].Topics[0].OPAConfigYAML, "resource: bundles/tenant-a/billing.tar.gz") {
		t.Fatalf("expected rendered OPA config to contain bundle resource, got %q", plan.Tenants[0].Topics[0].OPAConfigYAML)
	}
	if !strings.Contains(plan.Tenants[0].Topics[0].ConfigMapManifestYAML, "kind: ConfigMap") {
		t.Fatalf("expected config map manifest to be rendered, got %q", plan.Tenants[0].Topics[0].ConfigMapManifestYAML)
	}
	if !strings.Contains(plan.Tenants[0].Topics[0].ConfigMapManifestYAML, "opa-config.yaml: |") {
		t.Fatalf("expected config map manifest to inline opa-config.yaml contents")
	}
	if !strings.Contains(plan.Tenants[0].Topics[0].DeploymentManifestYAML, DefaultOPAImage) {
		t.Fatalf("expected deployment manifest to pin default OPA image")
	}
	if !strings.Contains(plan.Tenants[0].Topics[0].DeploymentManifestYAML, "configMap:") || !strings.Contains(plan.Tenants[0].Topics[0].DeploymentManifestYAML, "mountPath: /config") {
		t.Fatalf("expected deployment manifest to mount rendered config map, got %q", plan.Tenants[0].Topics[0].DeploymentManifestYAML)
	}
}

func TestBuildPlanUsesConfiguredOPAImage(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAImage:       "registry.example.com/opa:1.13.0",
		},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	if !strings.Contains(plan.Tenants[0].Topics[0].DeploymentManifestYAML, "registry.example.com/opa:1.13.0") {
		t.Fatalf("expected deployment manifest to use configured OPA image, got %q", plan.Tenants[0].Topics[0].DeploymentManifestYAML)
	}
	if strings.Contains(plan.Tenants[0].Topics[0].DeploymentManifestYAML, DefaultOPAImage) {
		t.Fatalf("expected deployment manifest to avoid fallback image when override is provided")
	}
}
