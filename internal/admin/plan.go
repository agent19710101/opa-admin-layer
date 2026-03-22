package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Plan struct {
	GeneratedAt string       `json:"generatedAt"`
	Name        string       `json:"name"`
	Topology    string       `json:"topology"`
	Tenants     []TenantPlan `json:"tenants"`
}

type TenantPlan struct {
	Name   string      `json:"name"`
	Topics []TopicPlan `json:"topics"`
}

type TopicPlan struct {
	Name                   string            `json:"name"`
	BundleURL              string            `json:"bundleURL"`
	DecisionPath           string            `json:"decisionPath"`
	ListenAddress          string            `json:"listenAddress"`
	Labels                 map[string]string `json:"labels,omitempty"`
	OPAConfigYAML          string            `json:"opaConfigYAML"`
	DeploymentManifestYAML string            `json:"deploymentManifestYAML"`
}

func BuildPlan(spec Specification) (Plan, error) {
	if issues := Validate(spec); len(issues) > 0 {
		return Plan{}, errors.New(strings.Join(issues, "; "))
	}
	normalized := normalize(spec)
	plan := Plan{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Name:        normalized.Name,
		Topology:    "opa-only",
		Tenants:     make([]TenantPlan, 0, len(normalized.Tenants)),
	}
	for _, tenant := range normalized.Tenants {
		tenantPlan := TenantPlan{Name: tenant.Name, Topics: make([]TopicPlan, 0, len(tenant.Topics))}
		for _, topic := range tenant.Topics {
			bundleURL := strings.TrimRight(normalized.ControlPlane.BaseServiceURL, "/") + "/" + strings.TrimLeft(topic.BundleResource, "/")
			tenantPlan.Topics = append(tenantPlan.Topics, TopicPlan{
				Name:                   topic.Name,
				BundleURL:              bundleURL,
				DecisionPath:           topic.DecisionPath,
				ListenAddress:          normalized.ControlPlane.DefaultListenAddress,
				Labels:                 topic.Labels,
				OPAConfigYAML:          renderOPAConfigYAML(normalized.ControlPlane.BaseServiceURL, topic.BundleResource),
				DeploymentManifestYAML: renderDeploymentYAML(normalized.Name, tenant.Name, topic.Name, normalized.ControlPlane.DefaultListenAddress, normalized.ControlPlane.OPAImage),
			})
		}
		plan.Tenants = append(plan.Tenants, tenantPlan)
	}
	return plan, nil
}

func renderOPAConfigYAML(baseURL, bundleResource string) string {
	return fmt.Sprintf(`services:
  controlplane:
    url: %s
bundles:
  tenant_topic:
    service: controlplane
    resource: %s
    persist: true
`, baseURL, bundleResource)
}

func renderDeploymentYAML(appName, tenantName, topicName, listenAddress, image string) string {
	name := fmt.Sprintf("%s-%s-%s-opa", sanitizeName(appName), sanitizeName(tenantName), sanitizeName(topicName))
	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
    app.kubernetes.io/name: %s
    app.kubernetes.io/component: opa
    app.kubernetes.io/tenant: %s
    app.kubernetes.io/topic: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: %s
  template:
    metadata:
      labels:
        app.kubernetes.io/name: %s
        app.kubernetes.io/component: opa
    spec:
      containers:
        - name: opa
          image: %s
          args:
            - run
            - --server
            - --addr=%s
            - --config-file=/config/opa-config.yaml
`, name, sanitizeName(appName), tenantName, topicName, name, name, image, listenAddress)
}

func sanitizeName(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer("_", "-", " ", "-")
	return replacer.Replace(value)
}

func MarshalPlan(plan Plan) ([]byte, error) {
	return json.MarshalIndent(plan, "", "  ")
}
