package admin

import (
	"strings"
	"testing"
)

func TestValidateRejectsInvalidBaseServiceURL(t *testing.T) {
	tests := map[string]string{
		"empty":             "",
		"relative":          "/control",
		"unsupportedScheme": "ftp://control.example.com",
		"missingHost":       "https:///bundles",
		"fragment":          "https://control.example.com/bundles#fragment",
	}

	for name, baseServiceURL := range tests {
		t.Run(name, func(t *testing.T) {
			spec := Specification{
				Name:         "demo",
				ControlPlane: ControlPlane{BaseServiceURL: baseServiceURL},
				Tenants: []Tenant{{
					Name:   "tenant-a",
					Topics: []Topic{{Name: "billing"}},
				}},
			}

			issues := Validate(spec)
			if len(issues) == 0 {
				t.Fatalf("expected invalid baseServiceURL issue")
			}
			if !strings.Contains(strings.Join(issues, "\n"), "controlPlane.baseServiceURL is invalid") {
				t.Fatalf("expected baseServiceURL validation issue, got %#v", issues)
			}
		})
	}
}

func TestValidateRejectsInvalidNamespace(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			Namespace:      "Team-A",
		},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}},
		}},
	}

	issues := Validate(spec)
	if len(issues) == 0 {
		t.Fatal("expected invalid namespace issue")
	}
	if !strings.Contains(strings.Join(issues, "\n"), "controlPlane.namespace is invalid") {
		t.Fatalf("expected namespace validation issue, got %#v", issues)
	}
}

func TestValidateAllowsExplicitListenAddressShapes(t *testing.T) {
	tests := map[string]string{
		"portOnly":     ":8181",
		"ipv4HostPort": "127.0.0.1:8282",
		"ipv6HostPort": "[::1]:8181",
	}

	for name, defaultListenAddress := range tests {
		t.Run(name, func(t *testing.T) {
			spec := Specification{
				Name: "demo",
				ControlPlane: ControlPlane{
					BaseServiceURL:       "https://control.example.com",
					DefaultListenAddress: defaultListenAddress,
				},
				Tenants: []Tenant{{
					Name:   "tenant-a",
					Topics: []Topic{{Name: "billing"}},
				}},
			}

			if issues := Validate(spec); len(issues) > 0 {
				t.Fatalf("expected listen address %q to pass validation, got %#v", defaultListenAddress, issues)
			}
		})
	}
}

func TestValidateRejectsInvalidDefaultListenAddress(t *testing.T) {
	tests := map[string]string{
		"missingPort":    "localhost",
		"barePort":       "8181",
		"nonnumericPort": ":abc",
		"outOfRangePort": "127.0.0.1:70000",
	}

	for name, defaultListenAddress := range tests {
		t.Run(name, func(t *testing.T) {
			spec := Specification{
				Name: "demo",
				ControlPlane: ControlPlane{
					BaseServiceURL:       "https://control.example.com",
					DefaultListenAddress: defaultListenAddress,
				},
				Tenants: []Tenant{{
					Name:   "tenant-a",
					Topics: []Topic{{Name: "billing"}},
				}},
			}

			issues := Validate(spec)
			if len(issues) == 0 {
				t.Fatalf("expected invalid default listen address issue for %q", defaultListenAddress)
			}
			if !strings.Contains(strings.Join(issues, "\n"), "controlPlane.defaultListenAddress is invalid") {
				t.Fatalf("expected defaultListenAddress validation issue, got %#v", issues)
			}
		})
	}
}

func TestParseListenAddressPortUsesSharedStrictContract(t *testing.T) {
	tests := map[string]struct {
		listenAddress string
		wantPort      int
		wantErr       string
	}{
		"defaultEmpty":      {listenAddress: "", wantPort: 8181},
		"defaultWhitespace": {listenAddress: "  ", wantPort: 8181},
		"portOnly":          {listenAddress: ":8181", wantPort: 8181},
		"hostPort":          {listenAddress: "127.0.0.1:8282", wantPort: 8282},
		"ipv6HostPort":      {listenAddress: "[::1]:8383", wantPort: 8383},
		"missingPort":       {listenAddress: "localhost", wantErr: "must use :port, host:port, or [ipv6]:port syntax"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseListenAddressPort(tc.listenAddress)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseListenAddressPort returned error: %v", err)
			}
			if got != tc.wantPort {
				t.Fatalf("port mismatch: got %d want %d", got, tc.wantPort)
			}
		})
	}
}

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
	if !strings.Contains(plan.Tenants[0].Topics[0].ServiceManifestYAML, "kind: Service") {
		t.Fatalf("expected service manifest to be rendered, got %q", plan.Tenants[0].Topics[0].ServiceManifestYAML)
	}
	if !strings.Contains(plan.Tenants[0].Topics[0].ServiceManifestYAML, "type: ClusterIP") {
		t.Fatalf("expected service manifest to default to ClusterIP, got %q", plan.Tenants[0].Topics[0].ServiceManifestYAML)
	}
	if strings.Contains(plan.Tenants[0].Topics[0].ServiceManifestYAML, "annotations:") {
		t.Fatalf("expected service manifest to omit annotations block by default, got %q", plan.Tenants[0].Topics[0].ServiceManifestYAML)
	}
	if !strings.Contains(plan.Tenants[0].Topics[0].ServiceManifestYAML, "port: 8181") || !strings.Contains(plan.Tenants[0].Topics[0].ServiceManifestYAML, "targetPort: 8181") {
		t.Fatalf("expected service manifest to use derived default port, got %q", plan.Tenants[0].Topics[0].ServiceManifestYAML)
	}
	for _, expected := range []string{
		"containerPort: 8181",
		"path: /health?plugins",
		"path: /health\n",
		"port: 8181",
	} {
		if !strings.Contains(plan.Tenants[0].Topics[0].DeploymentManifestYAML, expected) {
			t.Fatalf("expected deployment manifest to contain %q, got %q", expected, plan.Tenants[0].Topics[0].DeploymentManifestYAML)
		}
	}
	if strings.Contains(plan.Tenants[0].Topics[0].DeploymentManifestYAML, "resources:") {
		t.Fatalf("expected deployment manifest to omit resources block by default, got %q", plan.Tenants[0].Topics[0].DeploymentManifestYAML)
	}
}

func TestBuildPlanUsesConfiguredOPAResources(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{CPU: "100m", Memory: "128Mi"},
				Limits:   &ResourceList{Memory: "512Mi"},
			},
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
	deployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	for _, expected := range []string{
		"resources:",
		"requests:",
		`cpu: "100m"`,
		`memory: "128Mi"`,
		"limits:",
		`memory: "512Mi"`,
	} {
		if !strings.Contains(deployment, expected) {
			t.Fatalf("expected deployment manifest to contain %q, got %q", expected, deployment)
		}
	}
}

func TestBuildPlanMergesTopicOPAResourcesOverSharedDefaults(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{CPU: "100m", Memory: "128Mi"},
				Limits:   &ResourceList{Memory: "512Mi"},
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				OPAResources: ResourceRequirements{
					Requests: &ResourceList{Memory: "256Mi"},
					Limits:   &ResourceList{CPU: "750m"},
				},
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	deployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	for _, expected := range []string{
		"resources:",
		"requests:",
		`cpu: "100m"`,
		`memory: "256Mi"`,
		"limits:",
		`cpu: "750m"`,
		`memory: "512Mi"`,
	} {
		if !strings.Contains(deployment, expected) {
			t.Fatalf("expected deployment manifest to contain %q, got %q", expected, deployment)
		}
	}
	if strings.Contains(deployment, `memory: "128Mi"`) {
		t.Fatalf("expected topic request memory override to replace shared value, got %q", deployment)
	}
}

func TestBuildPlanUsesConfiguredServiceMetadata(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:        "https://control.example.com",
			Namespace:             "policy-system",
			ServiceType:           "LoadBalancer",
			ExternalTrafficPolicy: "Local",
			InternalTrafficPolicy: "Local",
			SessionAffinity:       "ClientIP",
			ServiceAnnotations: map[string]string{
				"service.beta.kubernetes.io/aws-load-balancer-scheme": "internal",
				"example.com/health-check-path":                       "/health?plugins",
			},
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
	configMap := plan.Tenants[0].Topics[0].ConfigMapManifestYAML
	deployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	service := plan.Tenants[0].Topics[0].ServiceManifestYAML
	if got := plan.Tenants[0].Topics[0].Namespace; got != "policy-system" {
		t.Fatalf("expected topic plan namespace to be recorded, got %q", got)
	}
	for _, manifest := range []string{configMap, deployment, service} {
		if !strings.Contains(manifest, "namespace: policy-system") {
			t.Fatalf("expected manifest to include configured namespace, got %q", manifest)
		}
	}
	if !strings.Contains(service, "type: LoadBalancer") {
		t.Fatalf("expected service manifest to use configured service type, got %q", service)
	}
	if strings.Contains(service, "type: ClusterIP") {
		t.Fatalf("expected configured service type to replace default ClusterIP")
	}
	if !strings.Contains(service, "annotations:") {
		t.Fatalf("expected service manifest to render annotations block, got %q", service)
	}
	if !strings.Contains(service, `service.beta.kubernetes.io/aws-load-balancer-scheme: "internal"`) {
		t.Fatalf("expected service manifest to include service annotation, got %q", service)
	}
	if !strings.Contains(service, `example.com/health-check-path: "/health?plugins"`) {
		t.Fatalf("expected service manifest to quote annotation values safely, got %q", service)
	}
	if !strings.Contains(service, "externalTrafficPolicy: Local") {
		t.Fatalf("expected service manifest to include configured externalTrafficPolicy, got %q", service)
	}
	if !strings.Contains(service, "internalTrafficPolicy: Local") {
		t.Fatalf("expected service manifest to include configured internalTrafficPolicy, got %q", service)
	}
	if !strings.Contains(service, "sessionAffinity: ClientIP") {
		t.Fatalf("expected service manifest to include configured sessionAffinity, got %q", service)
	}
}

func TestBuildPlanMergesTopicServiceOverridesOverSharedDefaults(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:        "https://control.example.com",
			ServiceType:           "LoadBalancer",
			ExternalTrafficPolicy: "Cluster",
			InternalTrafficPolicy: "Cluster",
			SessionAffinity:       "ClientIP",
			ServiceAnnotations: map[string]string{
				"example.com/health-check-path": "/health",
				"example.com/scope":             "shared",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:                  "billing",
				ServiceType:           "NodePort",
				ExternalTrafficPolicy: "Local",
				InternalTrafficPolicy: "Local",
				SessionAffinity:       "None",
				ServiceAnnotations: map[string]string{
					"example.com/scope":    "billing",
					"example.com/exposure": "public",
				},
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	service := plan.Tenants[0].Topics[0].ServiceManifestYAML
	for _, expected := range []string{
		"type: NodePort",
		"externalTrafficPolicy: Local",
		"internalTrafficPolicy: Local",
		"sessionAffinity: None",
		`example.com/health-check-path: "/health"`,
		`example.com/scope: "billing"`,
		`example.com/exposure: "public"`,
	} {
		if !strings.Contains(service, expected) {
			t.Fatalf("expected service manifest to contain %q, got %q", expected, service)
		}
	}
	if strings.Contains(service, "type: LoadBalancer") {
		t.Fatalf("expected topic service type override to replace shared value, got %q", service)
	}
	if strings.Contains(service, `example.com/scope: "shared"`) {
		t.Fatalf("expected topic annotation override to replace shared value, got %q", service)
	}
	if strings.Contains(service, "internalTrafficPolicy: Cluster") {
		t.Fatalf("expected topic internalTrafficPolicy override to replace shared value, got %q", service)
	}
	if strings.Contains(service, "sessionAffinity: ClientIP") {
		t.Fatalf("expected topic sessionAffinity override to replace shared value, got %q", service)
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

func TestBuildPlanRendersProbesUsingExplicitListenAddressPort(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:       "https://control.example.com",
			DefaultListenAddress: "127.0.0.1:8282",
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

	deployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	if !strings.Contains(deployment, "containerPort: 8282") {
		t.Fatalf("expected explicit listen port to render as containerPort, got %q", deployment)
	}
	if count := strings.Count(deployment, "port: 8282"); count != 2 {
		t.Fatalf("expected both probes to target explicit listen port, got count=%d manifest=%q", count, deployment)
	}

	service := plan.Tenants[0].Topics[0].ServiceManifestYAML
	if !strings.Contains(service, "port: 8282") || !strings.Contains(service, "targetPort: 8282") {
		t.Fatalf("expected service manifest to use explicit listen port, got %q", service)
	}
}

func TestBuildPlanPropagatesTopicLabelsIntoRenderedManifests(t *testing.T) {
	spec := Specification{
		Name:         "demo",
		ControlPlane: ControlPlane{BaseServiceURL: "https://control.example.com"},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				Labels: map[string]string{
					"environment":            "dev",
					"owner":                  "platform",
					"app.kubernetes.io/name": "overridden",
				},
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}

	configMap := plan.Tenants[0].Topics[0].ConfigMapManifestYAML
	deployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	service := plan.Tenants[0].Topics[0].ServiceManifestYAML

	for _, manifest := range []string{configMap, deployment, service} {
		if !strings.Contains(manifest, `environment: "dev"`) {
			t.Fatalf("expected propagated environment label in manifest, got %q", manifest)
		}
		if !strings.Contains(manifest, `owner: "platform"`) {
			t.Fatalf("expected propagated owner label in manifest, got %q", manifest)
		}
		if strings.Contains(manifest, `app.kubernetes.io/name: "overridden"`) {
			t.Fatalf("expected built-in app.kubernetes.io/name label to win over topic label override, got %q", manifest)
		}
	}

	if !strings.Contains(deployment, `app.kubernetes.io/name: "demo-tenant-a-billing-opa"`) {
		t.Fatalf("expected built-in deployment identity label to remain present, got %q", deployment)
	}
	if strings.Count(deployment, `environment: "dev"`) < 2 {
		t.Fatalf("expected propagated label in deployment metadata and pod template, got %q", deployment)
	}
	if !strings.Contains(service, `app.kubernetes.io/name: "demo-tenant-a-billing-opa"`) {
		t.Fatalf("expected service selector/labels to retain built-in identity, got %q", service)
	}
}

func TestValidateRejectsInvalidTopicLabelKeyAndValue(t *testing.T) {
	spec := Specification{
		Name:         "demo",
		ControlPlane: ControlPlane{BaseServiceURL: "https://control.example.com"},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				Labels: map[string]string{
					"Example.com/owner": "platform!",
				},
			}},
		}},
	}

	issues := Validate(spec)
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d: %#v", len(issues), issues)
	}
	if !strings.Contains(strings.Join(issues, "\n"), `label key "Example.com/owner" is invalid`) {
		t.Fatalf("expected invalid label key issue, got %#v", issues)
	}
	if !strings.Contains(strings.Join(issues, "\n"), `label "Example.com/owner" has invalid value "platform!"`) {
		t.Fatalf("expected invalid label value issue, got %#v", issues)
	}
}

func TestBuildPlanRejectsInvalidTopicLabels(t *testing.T) {
	spec := Specification{
		Name:         "demo",
		ControlPlane: ControlPlane{BaseServiceURL: "https://control.example.com"},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				Labels: map[string]string{
					"owner": "platform!",
				},
			}},
		}},
	}

	_, err := BuildPlan(spec)
	if err == nil {
		t.Fatal("expected BuildPlan to reject invalid topic labels")
	}
	if !strings.Contains(err.Error(), `label "owner" has invalid value "platform!"`) {
		t.Fatalf("expected invalid label error, got %v", err)
	}
}

func TestValidateRejectsRenderedResourceNamesWithInvalidCharacters(t *testing.T) {
	spec := Specification{
		Name:         "demo!",
		ControlPlane: ControlPlane{BaseServiceURL: "https://control.example.com"},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	if !strings.Contains(joined, `renders invalid deployment name "demo!-tenant-a-billing-opa"`) {
		t.Fatalf("expected invalid deployment-name issue, got %#v", issues)
	}
	if !strings.Contains(joined, `renders invalid configmap name "demo!-tenant-a-billing-opa-config"`) {
		t.Fatalf("expected invalid configmap-name issue, got %#v", issues)
	}
	if !strings.Contains(joined, `renders invalid service name "demo!-tenant-a-billing-opa"`) {
		t.Fatalf("expected invalid service-name issue, got %#v", issues)
	}
}

func TestValidateRejectsRenderedResourceNamesThatExceedLengthBudget(t *testing.T) {
	spec := Specification{
		Name:         strings.Repeat("a", 30),
		ControlPlane: ControlPlane{BaseServiceURL: "https://control.example.com"},
		Tenants: []Tenant{{
			Name:   strings.Repeat("b", 20),
			Topics: []Topic{{Name: strings.Repeat("c", 20)}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	if !strings.Contains(joined, `must be 63 characters or fewer`) {
		t.Fatalf("expected rendered-name length issue, got %#v", issues)
	}
	if !strings.Contains(joined, `renders invalid deployment name`) {
		t.Fatalf("expected deployment-name length issue, got %#v", issues)
	}
}

func TestValidateRejectsInvalidServiceTypeTrafficPolicyAnnotationKeyAndEmptyOPAResources(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:        "https://control.example.com",
			ServiceType:           "ExternalName",
			ExternalTrafficPolicy: "Edge",
			InternalTrafficPolicy: "Sideways",
			SessionAffinity:       "Sticky",
			ServiceAnnotations: map[string]string{
				"Example.com/internal": "true",
			},
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{},
			},
		},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	if !strings.Contains(joined, "controlPlane.serviceType is invalid") {
		t.Fatalf("expected invalid service type issue, got %#v", issues)
	}
	if !strings.Contains(joined, "ClusterIP, NodePort, or LoadBalancer") {
		t.Fatalf("expected allowed service types in issue, got %#v", issues)
	}
	if !strings.Contains(joined, `controlPlane.serviceAnnotations key "Example.com/internal" is invalid`) {
		t.Fatalf("expected invalid service annotation key issue, got %#v", issues)
	}
	if !strings.Contains(joined, "controlPlane.externalTrafficPolicy is invalid") {
		t.Fatalf("expected invalid external traffic policy issue, got %#v", issues)
	}
	if !strings.Contains(joined, "controlPlane.internalTrafficPolicy is invalid") {
		t.Fatalf("expected invalid internal traffic policy issue, got %#v", issues)
	}
	if !strings.Contains(joined, "controlPlane.sessionAffinity is invalid") {
		t.Fatalf("expected invalid session affinity issue, got %#v", issues)
	}
	if !strings.Contains(joined, "controlPlane.opaResources.requests must set cpu and/or memory") {
		t.Fatalf("expected empty opaResources requests issue, got %#v", issues)
	}
}

func TestValidateRejectsInvalidTopicServiceOverrides(t *testing.T) {
	spec := Specification{
		Name:         "demo",
		ControlPlane: ControlPlane{BaseServiceURL: "https://control.example.com"},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:                  "billing",
				ServiceType:           "ExternalName",
				ExternalTrafficPolicy: "Edge",
				InternalTrafficPolicy: "Sideways",
				SessionAffinity:       "Sticky",
				ServiceAnnotations: map[string]string{
					"Example.com/internal": "true",
				},
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	if !strings.Contains(joined, `tenant "tenant-a" topic "billing" serviceType is invalid`) {
		t.Fatalf("expected invalid topic service type issue, got %#v", issues)
	}
	if !strings.Contains(joined, `tenant "tenant-a" topic "billing" serviceAnnotations key "Example.com/internal" is invalid`) {
		t.Fatalf("expected invalid topic service annotation key issue, got %#v", issues)
	}
	if !strings.Contains(joined, `tenant "tenant-a" topic "billing" externalTrafficPolicy is invalid`) {
		t.Fatalf("expected invalid topic external traffic policy issue, got %#v", issues)
	}
	if !strings.Contains(joined, `tenant "tenant-a" topic "billing" internalTrafficPolicy is invalid`) {
		t.Fatalf("expected invalid topic internal traffic policy issue, got %#v", issues)
	}
	if !strings.Contains(joined, `tenant "tenant-a" topic "billing" sessionAffinity is invalid`) {
		t.Fatalf("expected invalid topic session affinity issue, got %#v", issues)
	}
}

func TestValidateRejectsExternalTrafficPolicyWithoutExternallyExposedService(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:        "https://control.example.com",
			ExternalTrafficPolicy: "Local",
		},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	if !strings.Contains(joined, `controlPlane.externalTrafficPolicy is invalid: requires serviceType NodePort or LoadBalancer, got ClusterIP`) {
		t.Fatalf("expected shared externalTrafficPolicy compatibility issue, got %#v", issues)
	}
}

func TestValidateRejectsInheritedExternalTrafficPolicyOnClusterIPTopic(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			ServiceType:    "NodePort",
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:                  "billing",
				ServiceType:           "ClusterIP",
				ExternalTrafficPolicy: "Local",
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	if !strings.Contains(joined, `tenant "tenant-a" topic "billing" effective externalTrafficPolicy is invalid: requires serviceType NodePort or LoadBalancer, got ClusterIP`) {
		t.Fatalf("expected effective topic externalTrafficPolicy compatibility issue, got %#v", issues)
	}
}

func TestValidateRejectsInvalidOPAResourceQuantities(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{CPU: "ten millicores", Memory: "128Mega"},
				Limits:   &ResourceList{Memory: "0x20"},
			},
		},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		"controlPlane.opaResources.requests.cpu is invalid",
		"controlPlane.opaResources.requests.memory is invalid",
		"controlPlane.opaResources.limits.memory is invalid",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected invalid quantity issue %q, got %#v", expected, issues)
		}
	}
}

func TestValidateRejectsEmptyAndInvalidTopicOPAResources(t *testing.T) {
	spec := Specification{
		Name:         "demo",
		ControlPlane: ControlPlane{BaseServiceURL: "https://control.example.com"},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				OPAResources: ResourceRequirements{
					Requests: &ResourceList{},
					Limits:   &ResourceList{Memory: "0x20"},
				},
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		`tenant "tenant-a" topic "billing" opaResources.requests must set cpu and/or memory`,
		`tenant "tenant-a" topic "billing" opaResources.limits.memory is invalid`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected topic opaResources issue %q, got %#v", expected, issues)
		}
	}
}

func TestValidateRejectsOPAResourceRequestsAboveLimits(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{CPU: "1000m", Memory: "512Mi"},
				Limits:   &ResourceList{CPU: "500m", Memory: "256Mi"},
			},
		},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		`controlPlane.opaResources.cpu request "1000m" must not exceed limit "500m"`,
		`controlPlane.opaResources.memory request "512Mi" must not exceed limit "256Mi"`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected opaResources budget issue %q, got %#v", expected, issues)
		}
	}
}

func TestValidateRejectsTopicOPAResourcesWhenMergedRequestsExceedInheritedLimits(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{CPU: "100m", Memory: "128Mi"},
				Limits:   &ResourceList{CPU: "500m", Memory: "256Mi"},
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				OPAResources: ResourceRequirements{
					Requests: &ResourceList{CPU: "750m", Memory: "512Mi"},
				},
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		`tenant "tenant-a" topic "billing" effective opaResources.cpu request "750m" must not exceed limit "500m"`,
		`tenant "tenant-a" topic "billing" effective opaResources.memory request "512Mi" must not exceed limit "256Mi"`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected merged topic opaResources budget issue %q, got %#v", expected, issues)
		}
	}
}

func TestBuildPlanMergesTopicDeploymentAndPodAnnotationsOverSharedDefaults(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			DeploymentAnnotations: map[string]string{
				"example.com/owner":           "platform",
				"example.com/revision-window": "shared",
			},
			PodAnnotations: map[string]string{
				"sidecar.istio.io/inject":    "false",
				"example.com/trace-sampling": "shared",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				DeploymentAnnotations: map[string]string{
					"example.com/revision-window": "billing",
					"example.com/rollout":         "canary",
				},
				PodAnnotations: map[string]string{
					"example.com/trace-sampling": "billing",
					"example.com/debug":          "enabled",
				},
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	deployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	for _, expected := range []string{
		`example.com/owner: "platform"`,
		`example.com/revision-window: "billing"`,
		`example.com/rollout: "canary"`,
		`sidecar.istio.io/inject: "false"`,
		`example.com/trace-sampling: "billing"`,
		`example.com/debug: "enabled"`,
	} {
		if !strings.Contains(deployment, expected) {
			t.Fatalf("expected deployment manifest to contain %q, got %q", expected, deployment)
		}
	}
	if strings.Contains(deployment, `example.com/revision-window: "shared"`) {
		t.Fatalf("expected topic deployment annotation override to replace shared value, got %q", deployment)
	}
	if strings.Contains(deployment, `example.com/trace-sampling: "shared"`) {
		t.Fatalf("expected topic pod annotation override to replace shared value, got %q", deployment)
	}
}

func TestValidateRejectsInvalidDeploymentAndPodAnnotationKeys(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			DeploymentAnnotations: map[string]string{
				"Example.com/deployment": "true",
			},
			PodAnnotations: map[string]string{
				"Example.com/shared": "true",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				DeploymentAnnotations: map[string]string{
					"Example.com/topic-deployment": "true",
				},
				PodAnnotations: map[string]string{
					"Example.com/topic": "true",
				},
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		`controlPlane.deploymentAnnotations key "Example.com/deployment" is invalid`,
		`controlPlane.podAnnotations key "Example.com/shared" is invalid`,
		`tenant "tenant-a" topic "billing" deploymentAnnotations key "Example.com/topic-deployment" is invalid`,
		`tenant "tenant-a" topic "billing" podAnnotations key "Example.com/topic" is invalid`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected invalid annotation key issue %q, got %#v", expected, issues)
		}
	}
}
