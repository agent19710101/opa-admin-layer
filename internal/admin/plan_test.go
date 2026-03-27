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

func TestValidateRejectsNegativeReplicas(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			Replicas:       -1,
		},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing", Replicas: -2}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		"controlPlane.replicas is invalid",
		`tenant "tenant-a" topic "billing" replicas is invalid`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected replica validation issue %q, got %#v", expected, issues)
		}
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
	if !strings.Contains(plan.Tenants[0].Topics[0].DeploymentManifestYAML, "replicas: 1") {
		t.Fatalf("expected deployment manifest to default replicas to 1, got %q", plan.Tenants[0].Topics[0].DeploymentManifestYAML)
	}
	if strings.Contains(plan.Tenants[0].Topics[0].DeploymentManifestYAML, "resources:") {
		t.Fatalf("expected deployment manifest to omit resources block by default, got %q", plan.Tenants[0].Topics[0].DeploymentManifestYAML)
	}
}

func TestBuildPlanUsesConfiguredReplicaOverrides(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			Replicas:       3,
		},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}, {Name: "support", Replicas: 5}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	billingDeployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	if !strings.Contains(billingDeployment, "replicas: 3") {
		t.Fatalf("expected shared replicas in billing deployment, got %q", billingDeployment)
	}
	supportDeployment := plan.Tenants[0].Topics[1].DeploymentManifestYAML
	if !strings.Contains(supportDeployment, "replicas: 5") {
		t.Fatalf("expected topic replica override in support deployment, got %q", supportDeployment)
	}
	if strings.Contains(supportDeployment, "replicas: 3") {
		t.Fatalf("expected topic replica override to replace shared value, got %q", supportDeployment)
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

func TestBuildPlanMergesTopicServiceLabelsOverSharedDefaultsWithoutMutatingOtherObjects(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			ServiceLabels: map[string]string{
				"example.com/service-scope": "shared",
				"example.com/team":          "platform",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				Labels: map[string]string{
					"example.com/team": "topic-metadata",
				},
				ServiceLabels: map[string]string{
					"example.com/service-scope": "billing",
					"example.com/ring":          "canary",
					"app.kubernetes.io/name":    "do-not-override",
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
		`example.com/service-scope: "billing"`,
		`example.com/ring: "canary"`,
		`example.com/team: "platform"`,
		`app.kubernetes.io/name: "demo-tenant-a-billing-opa"`,
	} {
		if !strings.Contains(service, expected) {
			t.Fatalf("expected service manifest to contain %q, got %q", expected, service)
		}
	}
	if strings.Contains(service, `app.kubernetes.io/name: "do-not-override"`) {
		t.Fatalf("expected built-in service label to remain immutable, got %q", service)
	}
	deployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	configMap := plan.Tenants[0].Topics[0].ConfigMapManifestYAML
	for _, manifest := range []string{deployment, configMap} {
		if strings.Contains(manifest, `example.com/service-scope`) || strings.Contains(manifest, `example.com/ring`) {
			t.Fatalf("expected service labels to stay service-scoped, got %q", manifest)
		}
	}
}

func TestValidateRejectsInvalidServiceLabels(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			ServiceLabels: map[string]string{
				"Example.com/shared": "ok",
				"example.com/value":  "bad!",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				ServiceLabels: map[string]string{
					"Example.com/topic": "ok",
					"example.com/ring":  "bad!",
				},
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		`controlPlane.serviceLabels key "Example.com/shared" is invalid`,
		`controlPlane.serviceLabels label "example.com/value" has invalid value "bad!"`,
		`tenant "tenant-a" topic "billing" serviceLabels key "Example.com/topic" is invalid`,
		`tenant "tenant-a" topic "billing" serviceLabels label "example.com/ring" has invalid value "bad!"`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected invalid service label issue %q, got %#v", expected, issues)
		}
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

func TestBuildPlanMergesTopicPodLabelsOverSharedDefaultsWithoutBreakingBuiltIns(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			PodLabels: map[string]string{
				"example.com/workload-class": "shared",
				"example.com/team":           "platform",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				Labels: map[string]string{
					"example.com/team": "topic-metadata",
				},
				PodLabels: map[string]string{
					"example.com/workload-class": "topic",
					"app.kubernetes.io/name":     "do-not-override",
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
		`example.com/workload-class: "topic"`,
		`example.com/team: "platform"`,
		`app.kubernetes.io/name: demo-tenant-a-billing-opa`,
	} {
		if !strings.Contains(deployment, expected) {
			t.Fatalf("expected deployment manifest to contain %q, got %q", expected, deployment)
		}
	}
	if strings.Contains(deployment, `app.kubernetes.io/name: "do-not-override"`) {
		t.Fatalf("expected built-in pod label to remain immutable, got %q", deployment)
	}
	service := plan.Tenants[0].Topics[0].ServiceManifestYAML
	configMap := plan.Tenants[0].Topics[0].ConfigMapManifestYAML
	for _, manifest := range []string{service, configMap} {
		if strings.Contains(manifest, `example.com/workload-class`) {
			t.Fatalf("expected pod labels to stay pod-scoped, got %q", manifest)
		}
	}
}

func TestValidateRejectsInvalidPodLabelKeysAndValues(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			PodLabels: map[string]string{
				"Example.com/shared": "ok",
				"example.com/value":  "bad!",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				PodLabels: map[string]string{
					"Example.com/topic": "ok",
					"example.com/track": "bad!",
				},
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		`controlPlane.podLabels key "Example.com/shared" is invalid`,
		`controlPlane.podLabels label "example.com/value" has invalid value "bad!"`,
		`tenant "tenant-a" topic "billing" podLabels key "Example.com/topic" is invalid`,
		`tenant "tenant-a" topic "billing" podLabels label "example.com/track" has invalid value "bad!"`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected invalid pod label issue %q, got %#v", expected, issues)
		}
	}
}

func TestBuildPlanRendersSharedConfigMapAnnotations(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			ConfigMapAnnotations: map[string]string{
				"reloader.stakater.com/match": "true",
				"example.com/source":          "generated",
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
	for _, expected := range []string{
		"annotations:",
		`reloader.stakater.com/match: "true"`,
		`example.com/source: "generated"`,
	} {
		if !strings.Contains(configMap, expected) {
			t.Fatalf("expected config map manifest to contain %q, got %q", expected, configMap)
		}
	}
}

func TestBuildPlanMergesTopicConfigMapAnnotationsOverSharedDefaults(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			ConfigMapAnnotations: map[string]string{
				"reloader.stakater.com/match": "true",
				"example.com/source":          "shared",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				ConfigMapAnnotations: map[string]string{
					"example.com/source": "billing",
					"example.com/team":   "payments",
				},
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	configMap := plan.Tenants[0].Topics[0].ConfigMapManifestYAML
	for _, expected := range []string{
		`reloader.stakater.com/match: "true"`,
		`example.com/source: "billing"`,
		`example.com/team: "payments"`,
	} {
		if !strings.Contains(configMap, expected) {
			t.Fatalf("expected config map manifest to contain %q, got %q", expected, configMap)
		}
	}
	if strings.Contains(configMap, `example.com/source: "shared"`) {
		t.Fatalf("expected topic config map annotation override to replace shared value, got %q", configMap)
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

func TestBuildPlanMergesTopicConfigMapLabelsOverSharedDefaultsWithoutMutatingOtherObjects(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			ConfigMapLabels: map[string]string{
				"example.com/config-scope": "shared",
				"example.com/team":         "platform",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				Labels: map[string]string{
					"example.com/team": "topic-metadata",
				},
				ConfigMapLabels: map[string]string{
					"example.com/config-scope": "billing",
					"example.com/ring":         "canary",
					"app.kubernetes.io/name":   "do-not-override",
				},
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	configMap := plan.Tenants[0].Topics[0].ConfigMapManifestYAML
	for _, expected := range []string{
		`example.com/config-scope: "billing"`,
		`example.com/ring: "canary"`,
		`example.com/team: "platform"`,
		`app.kubernetes.io/name: "demo-tenant-a-billing-opa"`,
	} {
		if !strings.Contains(configMap, expected) {
			t.Fatalf("expected config map manifest to contain %q, got %q", expected, configMap)
		}
	}
	if strings.Contains(configMap, `app.kubernetes.io/name: "do-not-override"`) {
		t.Fatalf("expected built-in config map label to remain immutable, got %q", configMap)
	}
	deployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	service := plan.Tenants[0].Topics[0].ServiceManifestYAML
	for _, manifest := range []string{deployment, service} {
		if strings.Contains(manifest, `example.com/config-scope`) || strings.Contains(manifest, `example.com/ring`) {
			t.Fatalf("expected config map labels to stay configmap-scoped, got %q", manifest)
		}
	}
}

func TestValidateRejectsInvalidConfigMapLabels(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			ConfigMapLabels: map[string]string{
				"Example.com/shared": "ok",
				"example.com/value":  "bad!",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				ConfigMapLabels: map[string]string{
					"Example.com/topic": "ok",
					"example.com/ring":  "bad!",
				},
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		`controlPlane.configMapLabels key "Example.com/shared" is invalid`,
		`controlPlane.configMapLabels label "example.com/value" has invalid value "bad!"`,
		`tenant "tenant-a" topic "billing" configMapLabels key "Example.com/topic" is invalid`,
		`tenant "tenant-a" topic "billing" configMapLabels label "example.com/ring" has invalid value "bad!"`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected invalid config map label issue %q, got %#v", expected, issues)
		}
	}
}

func TestBuildPlanAllowsTopicsToRemoveInheritedMetadata(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			ServiceAnnotations: map[string]string{
				"example.com/shared-service": "true",
			},
			ServiceLabels: map[string]string{
				"example.com/remove-service": "true",
			},
			ConfigMapAnnotations: map[string]string{
				"example.com/remove-config-annotation": "true",
			},
			ConfigMapLabels: map[string]string{
				"example.com/remove-config-label": "true",
			},
			DeploymentAnnotations: map[string]string{
				"example.com/remove-deployment-annotation": "true",
			},
			DeploymentLabels: map[string]string{
				"example.com/remove-deployment-label": "true",
			},
			PodAnnotations: map[string]string{
				"example.com/remove-pod-annotation": "true",
			},
			PodLabels: map[string]string{
				"example.com/remove-pod-label": "true",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:                     "billing",
				RemoveServiceAnnotations: []string{"example.com/shared-service"},
				RemoveServiceLabels:      []string{"example.com/remove-service"},
				RemoveConfigMapAnnotations: []string{
					"example.com/remove-config-annotation",
				},
				RemoveConfigMapLabels: []string{"example.com/remove-config-label"},
				RemoveDeploymentAnnotations: []string{
					"example.com/remove-deployment-annotation",
				},
				RemoveDeploymentLabels: []string{"example.com/remove-deployment-label"},
				RemovePodAnnotations:   []string{"example.com/remove-pod-annotation"},
				RemovePodLabels:        []string{"example.com/remove-pod-label"},
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	topicPlan := plan.Tenants[0].Topics[0]
	for _, check := range []struct {
		name     string
		manifest string
		removed  string
	}{
		{name: "service annotation", manifest: topicPlan.ServiceManifestYAML, removed: `example.com/shared-service`},
		{name: "service label", manifest: topicPlan.ServiceManifestYAML, removed: `example.com/remove-service`},
		{name: "config map annotation", manifest: topicPlan.ConfigMapManifestYAML, removed: `example.com/remove-config-annotation`},
		{name: "config map label", manifest: topicPlan.ConfigMapManifestYAML, removed: `example.com/remove-config-label`},
		{name: "deployment annotation", manifest: topicPlan.DeploymentManifestYAML, removed: `example.com/remove-deployment-annotation`},
		{name: "deployment label", manifest: topicPlan.DeploymentManifestYAML, removed: `example.com/remove-deployment-label`},
		{name: "pod annotation", manifest: topicPlan.DeploymentManifestYAML, removed: `example.com/remove-pod-annotation`},
		{name: "pod label", manifest: topicPlan.DeploymentManifestYAML, removed: `example.com/remove-pod-label`},
	} {
		if strings.Contains(check.manifest, check.removed) {
			t.Fatalf("expected %s to be removed, got %s", check.name, check.manifest)
		}
	}
	if !strings.Contains(topicPlan.ServiceManifestYAML, `app.kubernetes.io/name: "demo-tenant-a-billing-opa"`) {
		t.Fatalf("expected built-in service labels to remain present, got %s", topicPlan.ServiceManifestYAML)
	}
}

func TestValidateRejectsInvalidInheritedMetadataRemovalKeys(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:                       "billing",
				RemoveServiceAnnotations:   []string{"Example.com/service"},
				RemoveServiceLabels:        []string{"bad/key/format"},
				RemoveConfigMapAnnotations: []string{""},
				RemoveConfigMapLabels:      []string{"Example.com/config"},
				RemoveDeploymentAnnotations: []string{
					"Example.com/deployment",
				},
				RemoveDeploymentLabels: []string{"Example.com/deployment-label"},
				RemovePodAnnotations:   []string{"Example.com/pod"},
				RemovePodLabels:        []string{"Example.com/pod-label"},
			}},
		}},
	}

	issues := strings.Join(Validate(spec), "\n")
	for _, expected := range []string{
		`removeServiceAnnotations entry "Example.com/service" is invalid`,
		`removeServiceLabels entry "bad/key/format" is invalid`,
		`removeConfigMapAnnotations entry "" is invalid`,
		`removeConfigMapLabels entry "Example.com/config" is invalid`,
		`removeDeploymentAnnotations entry "Example.com/deployment" is invalid`,
		`removeDeploymentLabels entry "Example.com/deployment-label" is invalid`,
		`removePodAnnotations entry "Example.com/pod" is invalid`,
		`removePodLabels entry "Example.com/pod-label" is invalid`,
	} {
		if !strings.Contains(issues, expected) {
			t.Fatalf("expected invalid removal issue %q, got %s", expected, issues)
		}
	}
}

func TestValidateRejectsInvalidConfigMapDeploymentAndPodAnnotationKeys(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			ConfigMapAnnotations: map[string]string{
				"Example.com/config": "true",
			},
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
				ConfigMapAnnotations: map[string]string{
					"Example.com/topic-config": "true",
				},
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
		`controlPlane.configMapAnnotations key "Example.com/config" is invalid`,
		`controlPlane.deploymentAnnotations key "Example.com/deployment" is invalid`,
		`controlPlane.podAnnotations key "Example.com/shared" is invalid`,
		`tenant "tenant-a" topic "billing" configMapAnnotations key "Example.com/topic-config" is invalid`,
		`tenant "tenant-a" topic "billing" deploymentAnnotations key "Example.com/topic-deployment" is invalid`,
		`tenant "tenant-a" topic "billing" podAnnotations key "Example.com/topic" is invalid`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected invalid annotation key issue %q, got %#v", expected, issues)
		}
	}
}

func TestBuildPlanMergesTopicDeploymentLabelsOverSharedDefaultsWithoutMutatingOtherObjects(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			DeploymentLabels: map[string]string{
				"example.com/release-track": "shared",
				"example.com/team":          "platform",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				Labels: map[string]string{
					"example.com/team": "topic-metadata",
				},
				DeploymentLabels: map[string]string{
					"example.com/release-track": "billing",
					"example.com/ring":          "canary",
					"app.kubernetes.io/name":    "do-not-override",
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
		`example.com/release-track: "billing"`,
		`example.com/ring: "canary"`,
		`example.com/team: "platform"`,
		`app.kubernetes.io/name: "demo-tenant-a-billing-opa"`,
	} {
		if !strings.Contains(deployment, expected) {
			t.Fatalf("expected deployment manifest to contain %q, got %q", expected, deployment)
		}
	}
	if strings.Contains(deployment, `app.kubernetes.io/name: "do-not-override"`) {
		t.Fatalf("expected built-in deployment label to remain immutable, got %q", deployment)
	}
	service := plan.Tenants[0].Topics[0].ServiceManifestYAML
	configMap := plan.Tenants[0].Topics[0].ConfigMapManifestYAML
	for _, manifest := range []string{service, configMap} {
		if strings.Contains(manifest, `example.com/release-track`) || strings.Contains(manifest, `example.com/ring`) {
			t.Fatalf("expected deployment labels to stay deployment-scoped, got %q", manifest)
		}
	}
}

func TestValidateRejectsInvalidDeploymentLabelKeysAndValues(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			DeploymentLabels: map[string]string{
				"Example.com/shared": "ok",
				"example.com/value":  "bad!",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				DeploymentLabels: map[string]string{
					"Example.com/topic": "ok",
					"example.com/track": "bad!",
				},
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		`controlPlane.deploymentLabels key "Example.com/shared" is invalid`,
		`controlPlane.deploymentLabels label "example.com/value" has invalid value "bad!"`,
		`tenant "tenant-a" topic "billing" deploymentLabels key "Example.com/topic" is invalid`,
		`tenant "tenant-a" topic "billing" deploymentLabels label "example.com/track" has invalid value "bad!"`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected invalid deployment label issue %q, got %#v", expected, issues)
		}
	}
}

func TestBuildPlanMergesTopicServiceAccountNameOverSharedDefault(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:     "https://control.example.com",
			ServiceAccountName: "opa-shared",
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:               "billing",
				ServiceAccountName: "billing-opa",
			}, {
				Name: "support",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	billingTopic := plan.Tenants[0].Topics[0]
	billingDeployment := billingTopic.DeploymentManifestYAML
	if !strings.Contains(billingDeployment, "serviceAccountName: billing-opa") {
		t.Fatalf("expected topic serviceAccountName override in billing deployment, got %q", billingDeployment)
	}
	if strings.Contains(billingDeployment, "serviceAccountName: opa-shared") {
		t.Fatalf("expected topic serviceAccountName override to replace shared value, got %q", billingDeployment)
	}
	if !strings.Contains(billingTopic.ServiceAccountManifestYAML, "kind: ServiceAccount") || !strings.Contains(billingTopic.ServiceAccountManifestYAML, "name: billing-opa") {
		t.Fatalf("expected billing topic to render topic-scoped service account manifest, got %q", billingTopic.ServiceAccountManifestYAML)
	}
	supportTopic := plan.Tenants[0].Topics[1]
	supportDeployment := supportTopic.DeploymentManifestYAML
	if !strings.Contains(supportDeployment, "serviceAccountName: opa-shared") {
		t.Fatalf("expected support deployment to inherit shared serviceAccountName, got %q", supportDeployment)
	}
	if !strings.Contains(supportTopic.ServiceAccountManifestYAML, "name: opa-shared") {
		t.Fatalf("expected support topic to render inherited shared service account manifest, got %q", supportTopic.ServiceAccountManifestYAML)
	}
}

func TestBuildPlanOmitsServiceAccountManifestByDefault(t *testing.T) {
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
	topic := plan.Tenants[0].Topics[0]
	if topic.ServiceAccountManifestYAML != "" {
		t.Fatalf("expected service account manifest to be omitted by default, got %q", topic.ServiceAccountManifestYAML)
	}
	if strings.Contains(topic.DeploymentManifestYAML, "serviceAccountName:") {
		t.Fatalf("expected deployment manifest to omit serviceAccountName by default, got %q", topic.DeploymentManifestYAML)
	}
}

func TestBuildPlanMergesTopicServiceAccountAnnotationsOverSharedDefault(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:     "https://control.example.com",
			ServiceAccountName: "opa-shared",
			ServiceAccountAnnotations: map[string]string{
				"eks.amazonaws.com/role-arn": "arn:aws:iam::123456789012:role/shared-opa",
				"example.com/source":         "shared",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:               "billing",
				ServiceAccountName: "billing-opa",
				ServiceAccountAnnotations: map[string]string{
					"example.com/source": "billing",
					"example.com/team":   "payments",
				},
				RemoveServiceAccountAnnotations: []string{"eks.amazonaws.com/role-arn"},
			}, {
				Name: "support",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	billingServiceAccount := plan.Tenants[0].Topics[0].ServiceAccountManifestYAML
	for _, expected := range []string{
		`example.com/source: "billing"`,
		`example.com/team: "payments"`,
	} {
		if !strings.Contains(billingServiceAccount, expected) {
			t.Fatalf("expected billing service account manifest to contain %q, got %q", expected, billingServiceAccount)
		}
	}
	if strings.Contains(billingServiceAccount, `eks.amazonaws.com/role-arn`) {
		t.Fatalf("expected billing service account manifest to remove inherited annotation, got %q", billingServiceAccount)
	}
	supportServiceAccount := plan.Tenants[0].Topics[1].ServiceAccountManifestYAML
	for _, expected := range []string{
		`eks.amazonaws.com/role-arn: "arn:aws:iam::123456789012:role/shared-opa"`,
		`example.com/source: "shared"`,
	} {
		if !strings.Contains(supportServiceAccount, expected) {
			t.Fatalf("expected support service account manifest to contain %q, got %q", expected, supportServiceAccount)
		}
	}
}

func TestBuildPlanMergesTopicServiceAccountLabelsOverSharedDefault(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:     "https://control.example.com",
			ServiceAccountName: "opa-shared",
			ServiceAccountLabels: map[string]string{
				"example.com/service-account-scope": "shared",
				"example.com/team":                  "platform",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:               "billing",
				ServiceAccountName: "billing-opa",
				Labels: map[string]string{
					"example.com/workload": "billing",
				},
				ServiceAccountLabels: map[string]string{
					"example.com/service-account-scope": "billing",
					"example.com/ring":                  "canary",
					"app.kubernetes.io/name":            "ignored",
				},
				RemoveServiceAccountLabels: []string{"example.com/team"},
			}, {
				Name: "support",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	billingServiceAccount := plan.Tenants[0].Topics[0].ServiceAccountManifestYAML
	for _, expected := range []string{
		`labels:`,
		`app.kubernetes.io/name: "demo-tenant-a-billing-opa"`,
		`app.kubernetes.io/component: "opa"`,
		`example.com/workload: "billing"`,
		`example.com/service-account-scope: "billing"`,
		`example.com/ring: "canary"`,
	} {
		if !strings.Contains(billingServiceAccount, expected) {
			t.Fatalf("expected billing service account manifest to contain %q, got %q", expected, billingServiceAccount)
		}
	}
	if strings.Contains(billingServiceAccount, `example.com/team`) {
		t.Fatalf("expected billing service account manifest to remove inherited label, got %q", billingServiceAccount)
	}
	if strings.Contains(billingServiceAccount, `app.kubernetes.io/name: "ignored"`) {
		t.Fatalf("expected built-in service account labels to remain immutable, got %q", billingServiceAccount)
	}

	supportServiceAccount := plan.Tenants[0].Topics[1].ServiceAccountManifestYAML
	for _, expected := range []string{
		`app.kubernetes.io/name: "demo-tenant-a-support-opa"`,
		`example.com/service-account-scope: "shared"`,
		`example.com/team: "platform"`,
	} {
		if !strings.Contains(supportServiceAccount, expected) {
			t.Fatalf("expected support service account manifest to contain %q, got %q", expected, supportServiceAccount)
		}
	}
}

func TestBuildPlanMergesTopicAutomountServiceAccountTokenOverSharedDefault(t *testing.T) {
	sharedAutomount := false
	topicAutomount := true
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:               "https://control.example.com",
			ServiceAccountName:           "opa-shared",
			AutomountServiceAccountToken: &sharedAutomount,
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:                         "billing",
				ServiceAccountName:           "billing-opa",
				AutomountServiceAccountToken: &topicAutomount,
			}, {
				Name: "support",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	billingDeployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	if !strings.Contains(billingDeployment, "automountServiceAccountToken: true") {
		t.Fatalf("expected topic automountServiceAccountToken override in billing deployment, got %q", billingDeployment)
	}
	if strings.Contains(billingDeployment, "automountServiceAccountToken: false") {
		t.Fatalf("expected topic automountServiceAccountToken override to replace shared value, got %q", billingDeployment)
	}
	supportDeployment := plan.Tenants[0].Topics[1].DeploymentManifestYAML
	if !strings.Contains(supportDeployment, "automountServiceAccountToken: false") {
		t.Fatalf("expected support deployment to inherit shared automountServiceAccountToken, got %q", supportDeployment)
	}
}

func TestValidateRejectsInvalidServiceAccountName(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:     "https://control.example.com",
			ServiceAccountName: "OPA.Shared",
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:               "billing",
				ServiceAccountName: "billing_opa",
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		"controlPlane.serviceAccountName is invalid",
		`tenant "tenant-a" topic "billing" serviceAccountName is invalid`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected invalid serviceAccountName issue %q, got %#v", expected, issues)
		}
	}
}

func TestValidateAllowsRepeatedEffectiveServiceAccountNameForSharedBindings(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:     "https://control.example.com",
			ServiceAccountName: "opa-shared",
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
			}, {
				Name: "support",
			}},
		}},
	}

	issues := strings.Join(Validate(spec), "\n")
	if strings.Contains(issues, `effective serviceAccountName`) {
		t.Fatalf("expected repeated serviceAccountName values to be allowed for shared bindings, got %s", issues)
	}
}

func TestBuildPlanOmitsRenderedServiceAccountManifestForSharedBindings(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:     "https://control.example.com",
			ServiceAccountName: "opa-shared",
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
			}, {
				Name: "support",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	for _, topic := range plan.Tenants[0].Topics {
		if topic.ServiceAccountManifestYAML != "" {
			t.Fatalf("expected shared binding topic %q to omit rendered ServiceAccount manifest, got %q", topic.Name, topic.ServiceAccountManifestYAML)
		}
		if !strings.Contains(topic.DeploymentManifestYAML, "serviceAccountName: opa-shared") {
			t.Fatalf("expected deployment for topic %q to keep serviceAccountName binding, got %q", topic.Name, topic.DeploymentManifestYAML)
		}
	}
}

func TestValidateRejectsInvalidServiceAccountAnnotations(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			ServiceAccountAnnotations: map[string]string{
				"Example.com/shared": "true",
			},
			ServiceAccountLabels: map[string]string{
				"bad key": "shared!",
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				ServiceAccountAnnotations: map[string]string{
					"Example.com/topic": "true",
				},
				RemoveServiceAccountAnnotations: []string{"bad key"},
				ServiceAccountLabels: map[string]string{
					"Example.com/topic": "bad!",
				},
				RemoveServiceAccountLabels: []string{"also bad"},
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		`controlPlane.serviceAccountAnnotations key "Example.com/shared" is invalid`,
		`controlPlane.serviceAccountLabels key "bad key" is invalid`,
		`controlPlane.serviceAccountLabels label "bad key" has invalid value "shared!"`,
		`tenant "tenant-a" topic "billing" serviceAccountAnnotations key "Example.com/topic" is invalid`,
		`tenant "tenant-a" topic "billing" removeServiceAccountAnnotations entry "bad key" is invalid`,
		`tenant "tenant-a" topic "billing" serviceAccountLabels key "Example.com/topic" is invalid`,
		`tenant "tenant-a" topic "billing" serviceAccountLabels label "Example.com/topic" has invalid value "bad!"`,
		`tenant "tenant-a" topic "billing" removeServiceAccountLabels entry "also bad" is invalid`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected invalid serviceAccountAnnotations issue %q, got %#v", expected, issues)
		}
	}
}

func TestBuildPlanMergesTopicImagePullPolicyOverSharedDefault(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:  "https://control.example.com",
			ImagePullPolicy: "IfNotPresent",
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:            "billing",
				ImagePullPolicy: "Always",
			}, {
				Name: "support",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	billingDeployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	if !strings.Contains(billingDeployment, "imagePullPolicy: Always") {
		t.Fatalf("expected topic imagePullPolicy override in billing deployment, got %q", billingDeployment)
	}
	if strings.Contains(billingDeployment, "imagePullPolicy: IfNotPresent") {
		t.Fatalf("expected topic imagePullPolicy override to replace shared value, got %q", billingDeployment)
	}
	supportDeployment := plan.Tenants[0].Topics[1].DeploymentManifestYAML
	if !strings.Contains(supportDeployment, "imagePullPolicy: IfNotPresent") {
		t.Fatalf("expected support deployment to inherit shared imagePullPolicy, got %q", supportDeployment)
	}
}

func TestBuildPlanOmitsImagePullPolicyByDefault(t *testing.T) {
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
	deployment := plan.Tenants[0].Topics[0].DeploymentManifestYAML
	if strings.Contains(deployment, "imagePullPolicy:") {
		t.Fatalf("expected deployment manifest to omit imagePullPolicy by default, got %q", deployment)
	}
}

func TestValidateRejectsInvalidImagePullPolicy(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:  "https://control.example.com",
			ImagePullPolicy: "Sometimes",
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:            "billing",
				ImagePullPolicy: "OnDemand",
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		"controlPlane.imagePullPolicy is invalid",
		`tenant "tenant-a" topic "billing" imagePullPolicy is invalid`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected invalid imagePullPolicy issue %q, got %#v", expected, issues)
		}
	}
}

func TestBuildPlanRendersHPAFromInheritedAutoscaling(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{CPU: "100m"},
			},
			Autoscaling: &Autoscaling{
				MinReplicas:                    2,
				MaxReplicas:                    6,
				TargetCPUUtilizationPercentage: 70,
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	topic := plan.Tenants[0].Topics[0]
	if !strings.Contains(topic.DeploymentManifestYAML, "replicas: 2") {
		t.Fatalf("expected deployment replicas to follow autoscaling minReplicas, got %q", topic.DeploymentManifestYAML)
	}
	for _, expected := range []string{
		"kind: HorizontalPodAutoscaler",
		"name: demo-tenant-a-billing",
		"minReplicas: 2",
		"maxReplicas: 6",
		"averageUtilization: 70",
	} {
		if !strings.Contains(topic.HPAManifestYAML, expected) {
			t.Fatalf("expected HPA manifest to contain %q, got %q", expected, topic.HPAManifestYAML)
		}
	}
}

func TestValidateRejectsAutoscalingWithoutEffectiveCPURequest(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			Autoscaling: &Autoscaling{
				MinReplicas:                    2,
				MaxReplicas:                    5,
				TargetCPUUtilizationPercentage: 70,
			},
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{Memory: "128Mi"},
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	expected := `tenant "tenant-a" topic "billing" effective autoscaling requires effective opaResources.requests.cpu to be set for CPU utilization metrics`
	if !strings.Contains(joined, expected) {
		t.Fatalf("expected autoscaling cpu request issue %q, got %#v", expected, issues)
	}
}

func TestValidateAllowsAutoscalingWhenTopicSuppliesCPURequest(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{Memory: "128Mi"},
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
				Autoscaling: &Autoscaling{
					MinReplicas:                    2,
					MaxReplicas:                    5,
					TargetCPUUtilizationPercentage: 70,
				},
				OPAResources: ResourceRequirements{
					Requests: &ResourceList{CPU: "100m"},
				},
			}},
		}},
	}

	issues := Validate(spec)
	for _, issue := range issues {
		if strings.Contains(issue, "effective autoscaling requires effective opaResources.requests.cpu") {
			t.Fatalf("did not expect autoscaling cpu request issue, got %#v", issues)
		}
	}
}

func TestValidateRejectsConflictingOrInvalidAutoscaling(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			Replicas:       2,
			Autoscaling: &Autoscaling{
				MinReplicas:                    3,
				MaxReplicas:                    1,
				TargetCPUUtilizationPercentage: 101,
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:     "billing",
				Replicas: 4,
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	for _, expected := range []string{
		"controlPlane.autoscaling.maxReplicas must be greater than or equal to minReplicas",
		"controlPlane.autoscaling.targetCPUUtilizationPercentage must be between 1 and 100",
		"controlPlane.replicas is invalid: cannot be set when controlPlane.autoscaling is configured",
		`tenant "tenant-a" topic "billing" replicas is invalid: cannot be set when controlPlane.autoscaling is configured`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected autoscaling issue %q, got %#v", expected, issues)
		}
	}
}

func TestValidateRejectsAutoscalingWithoutEffectiveMemoryRequest(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{CPU: "100m"},
			},
			Autoscaling: &Autoscaling{
				MinReplicas:                       2,
				MaxReplicas:                       5,
				TargetMemoryUtilizationPercentage: 75,
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
			}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	expected := `tenant "tenant-a" topic "billing" effective autoscaling requires effective opaResources.requests.memory to be set for memory utilization metrics`
	if !strings.Contains(joined, expected) {
		t.Fatalf("expected autoscaling memory request issue %q, got %#v", expected, issues)
	}
}

func TestValidateRejectsAutoscalingWithoutConfiguredMetricTargets(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			Autoscaling: &Autoscaling{
				MinReplicas: 2,
				MaxReplicas: 5,
			},
		},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}},
		}},
	}

	issues := Validate(spec)
	joined := strings.Join(issues, "\n")
	if !strings.Contains(joined, "controlPlane.autoscaling must set targetCPUUtilizationPercentage and/or targetMemoryUtilizationPercentage") {
		t.Fatalf("expected missing autoscaling metrics issue, got %#v", issues)
	}
}

func TestBuildPlanRendersMemoryAutoscalingMetric(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{Memory: "256Mi"},
			},
			Autoscaling: &Autoscaling{
				MinReplicas:                       2,
				MaxReplicas:                       6,
				TargetMemoryUtilizationPercentage: 80,
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	hpa := plan.Tenants[0].Topics[0].HPAManifestYAML
	for _, expected := range []string{
		"kind: HorizontalPodAutoscaler",
		"name: memory",
		"averageUtilization: 80",
	} {
		if !strings.Contains(hpa, expected) {
			t.Fatalf("expected memory autoscaling manifest to contain %q, got %q", expected, hpa)
		}
	}
	if strings.Contains(hpa, "name: cpu") {
		t.Fatalf("expected memory-only autoscaling manifest to omit cpu metric, got %q", hpa)
	}
}

func TestBuildPlanRendersCPUAndMemoryAutoscalingMetricsInStableOrder(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{CPU: "100m", Memory: "256Mi"},
			},
			Autoscaling: &Autoscaling{
				MinReplicas:                       2,
				MaxReplicas:                       6,
				TargetCPUUtilizationPercentage:    70,
				TargetMemoryUtilizationPercentage: 80,
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	hpa := plan.Tenants[0].Topics[0].HPAManifestYAML
	cpuIndex := strings.Index(hpa, "name: cpu")
	memoryIndex := strings.Index(hpa, "name: memory")
	if cpuIndex == -1 || memoryIndex == -1 {
		t.Fatalf("expected cpu and memory metrics in autoscaling manifest, got %q", hpa)
	}
	if cpuIndex > memoryIndex {
		t.Fatalf("expected cpu metric before memory metric, got %q", hpa)
	}
}

func TestValidateRejectsInvalidAutoscalingBehavior(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{CPU: "100m"},
			},
			Autoscaling: &Autoscaling{
				MinReplicas:                    2,
				MaxReplicas:                    5,
				TargetCPUUtilizationPercentage: 70,
				Behavior:                       &AutoscalingBehavior{},
			},
		},
		Tenants: []Tenant{{
			Name:   "tenant-a",
			Topics: []Topic{{Name: "billing"}},
		}},
	}

	issues := strings.Join(Validate(spec), "\n")
	if !strings.Contains(issues, "controlPlane.autoscaling.behavior must configure scaleUp and/or scaleDown") {
		t.Fatalf("expected invalid autoscaling behavior issue, got %#v", issues)
	}

	spec.ControlPlane.Autoscaling.Behavior = &AutoscalingBehavior{
		ScaleUp: &AutoscalingBehaviorPolicy{},
	}
	issues = strings.Join(Validate(spec), "\n")
	if !strings.Contains(issues, "controlPlane.autoscaling.behavior.scaleUp must set stabilizationWindowSeconds, selectPolicy, and/or policies") {
		t.Fatalf("expected empty autoscaling policy issue, got %#v", issues)
	}

	window := -1
	spec.ControlPlane.Autoscaling.Behavior = &AutoscalingBehavior{
		ScaleUp: &AutoscalingBehaviorPolicy{StabilizationWindowSeconds: &window},
	}
	issues = strings.Join(Validate(spec), "\n")
	if !strings.Contains(issues, "controlPlane.autoscaling.behavior.scaleUp.stabilizationWindowSeconds must be zero or greater") {
		t.Fatalf("expected invalid autoscaling stabilization window issue, got %#v", issues)
	}

	spec.ControlPlane.Autoscaling.Behavior = &AutoscalingBehavior{
		ScaleDown: &AutoscalingBehaviorPolicy{
			SelectPolicy: "median",
			Policies:     []HPAScalingPolicy{{Type: "Requests", Value: 0, PeriodSeconds: 1900}},
		},
	}
	issues = strings.Join(Validate(spec), "\n")
	for _, expected := range []string{
		"controlPlane.autoscaling.behavior.scaleDown.selectPolicy is invalid: must be Max, Min, or Disabled",
		"controlPlane.autoscaling.behavior.scaleDown.policies[0].type must be Pods or Percent",
		"controlPlane.autoscaling.behavior.scaleDown.policies[0].value must be greater than zero",
		"controlPlane.autoscaling.behavior.scaleDown.policies[0].periodSeconds must be 1800 or fewer",
	} {
		if !strings.Contains(issues, expected) {
			t.Fatalf("expected invalid autoscaling behavior policy issue %q, got %#v", expected, issues)
		}
	}
}

func TestBuildPlanRendersAutoscalingBehavior(t *testing.T) {
	scaleUpWindow := 30
	scaleDownWindow := 300
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
			OPAResources: ResourceRequirements{
				Requests: &ResourceList{CPU: "100m", Memory: "256Mi"},
			},
			Autoscaling: &Autoscaling{
				MinReplicas:                       2,
				MaxReplicas:                       6,
				TargetCPUUtilizationPercentage:    70,
				TargetMemoryUtilizationPercentage: 80,
				Behavior: &AutoscalingBehavior{
					ScaleUp: &AutoscalingBehaviorPolicy{
						StabilizationWindowSeconds: &scaleUpWindow,
						SelectPolicy:               "Max",
						Policies:                   []HPAScalingPolicy{{Type: "Pods", Value: 2, PeriodSeconds: 60}},
					},
					ScaleDown: &AutoscalingBehaviorPolicy{
						StabilizationWindowSeconds: &scaleDownWindow,
						SelectPolicy:               "Min",
						Policies:                   []HPAScalingPolicy{{Type: "Percent", Value: 25, PeriodSeconds: 120}},
					},
				},
			},
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name: "billing",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	hpa := plan.Tenants[0].Topics[0].HPAManifestYAML
	for _, expected := range []string{
		"behavior:",
		"scaleUp:",
		"stabilizationWindowSeconds: 30",
		"selectPolicy: Max",
		"- type: Pods",
		"value: 2",
		"periodSeconds: 60",
		"scaleDown:",
		"stabilizationWindowSeconds: 300",
		"selectPolicy: Min",
		"- type: Percent",
		"value: 25",
		"periodSeconds: 120",
	} {
		if !strings.Contains(hpa, expected) {
			t.Fatalf("expected autoscaling behavior manifest to contain %q, got %q", expected, hpa)
		}
	}
}

func TestBuildPlanUsesTopicListenAddressOverride(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL:       "https://control.example.com",
			DefaultListenAddress: ":8181",
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:          "billing",
				ListenAddress: "127.0.0.1:8282",
			}, {
				Name: "support",
			}},
		}},
	}

	plan, err := BuildPlan(spec)
	if err != nil {
		t.Fatalf("BuildPlan returned error: %v", err)
	}
	billing := plan.Tenants[0].Topics[0]
	if billing.ListenAddress != "127.0.0.1:8282" {
		t.Fatalf("expected topic listenAddress override to be recorded, got %q", billing.ListenAddress)
	}
	for _, expected := range []string{
		"containerPort: 8282",
		"port: 8282",
		"targetPort: 8282",
		"--addr=127.0.0.1:8282",
	} {
		if !strings.Contains(billing.DeploymentManifestYAML+billing.ServiceManifestYAML, expected) {
			t.Fatalf("expected topic listenAddress override output %q, got deployment=%q service=%q", expected, billing.DeploymentManifestYAML, billing.ServiceManifestYAML)
		}
	}
	support := plan.Tenants[0].Topics[1]
	if support.ListenAddress != ":8181" {
		t.Fatalf("expected sibling topic to inherit control-plane listenAddress, got %q", support.ListenAddress)
	}
}

func TestValidateRejectsInvalidTopicListenAddress(t *testing.T) {
	spec := Specification{
		Name: "demo",
		ControlPlane: ControlPlane{
			BaseServiceURL: "https://control.example.com",
		},
		Tenants: []Tenant{{
			Name: "tenant-a",
			Topics: []Topic{{
				Name:          "billing",
				ListenAddress: "localhost",
			}},
		}},
	}

	issues := strings.Join(Validate(spec), "\n")
	if !strings.Contains(issues, `tenant "tenant-a" topic "billing" listenAddress is invalid`) {
		t.Fatalf("expected invalid topic listenAddress issue, got %s", issues)
	}
}
