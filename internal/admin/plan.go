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
	ConfigMapManifestYAML  string            `json:"configMapManifestYAML"`
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
			opaConfigYAML := renderOPAConfigYAML(normalized.ControlPlane.BaseServiceURL, topic.BundleResource)
			configMapName := topicConfigMapName(normalized.Name, tenant.Name, topic.Name)
			tenantPlan.Topics = append(tenantPlan.Topics, TopicPlan{
				Name:                   topic.Name,
				BundleURL:              bundleURL,
				DecisionPath:           topic.DecisionPath,
				ListenAddress:          normalized.ControlPlane.DefaultListenAddress,
				Labels:                 topic.Labels,
				OPAConfigYAML:          opaConfigYAML,
				ConfigMapManifestYAML:  renderConfigMapYAML(normalized.Name, tenant.Name, topic.Name, opaConfigYAML),
				DeploymentManifestYAML: renderDeploymentYAML(normalized.Name, tenant.Name, topic.Name, normalized.ControlPlane.DefaultListenAddress, normalized.ControlPlane.OPAImage, configMapName),
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

func renderConfigMapYAML(appName, tenantName, topicName, opaConfig string) string {
	name := topicConfigMapName(appName, tenantName, topicName)
	indentedConfig := strings.ReplaceAll(strings.TrimRight(opaConfig, "\n"), "\n", "\n    ")
	return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  labels:
    app.kubernetes.io/name: %s
    app.kubernetes.io/component: opa
    app.kubernetes.io/tenant: %s
    app.kubernetes.io/topic: %s
data:
  opa-config.yaml: |
    %s
`, name, sanitizeName(appName), tenantName, topicName, indentedConfig)
}

func renderDeploymentYAML(appName, tenantName, topicName, listenAddress, image, configMapName string) string {
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
      volumes:
        - name: opa-config
          configMap:
            name: %s
      containers:
        - name: opa
          image: %s
          volumeMounts:
            - name: opa-config
              mountPath: /config
              readOnly: true
          args:
            - run
            - --server
            - --addr=%s
            - --config-file=/config/opa-config.yaml
`, name, sanitizeName(appName), tenantName, topicName, name, name, configMapName, image, listenAddress)
}

func topicConfigMapName(appName, tenantName, topicName string) string {
	return fmt.Sprintf("%s-%s-%s-opa-config", sanitizeName(appName), sanitizeName(tenantName), sanitizeName(topicName))
}

func sanitizeName(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer("_", "-", " ", "-")
	return replacer.Replace(value)
}

func MarshalPlan(plan Plan) ([]byte, error) {
	return json.MarshalIndent(plan, "", "  ")
}
