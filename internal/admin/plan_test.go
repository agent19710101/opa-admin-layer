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
	if !strings.Contains(plan.Tenants[0].Topics[0].DeploymentManifestYAML, "openpolicyagent/opa:1.12.1") {
		t.Fatalf("expected deployment manifest to pin OPA image")
	}
}
