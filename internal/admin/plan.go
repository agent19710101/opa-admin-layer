package admin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
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
			builtInLabels := builtInTopicLabels(normalized.Name, tenant.Name, topic.Name)
			renderedLabels := mergeTopicLabels(builtInLabels, topic.Labels)
			configMapName := topicConfigMapName(normalized.Name, tenant.Name, topic.Name)
			tenantPlan.Topics = append(tenantPlan.Topics, TopicPlan{
				Name:                   topic.Name,
				BundleURL:              bundleURL,
				DecisionPath:           topic.DecisionPath,
				ListenAddress:          normalized.ControlPlane.DefaultListenAddress,
				Labels:                 topic.Labels,
				OPAConfigYAML:          opaConfigYAML,
				ConfigMapManifestYAML:  renderConfigMapYAML(configMapName, opaConfigYAML, renderedLabels),
				DeploymentManifestYAML: renderDeploymentYAML(deploymentName(normalized.Name, tenant.Name, topic.Name), normalized.ControlPlane.DefaultListenAddress, normalized.ControlPlane.OPAImage, configMapName, renderedLabels),
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

func renderConfigMapYAML(name, opaConfig string, labels map[string]string) string {
	indentedConfig := strings.ReplaceAll(strings.TrimRight(opaConfig, "\n"), "\n", "\n    ")
	return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  labels:
%sdata:
  opa-config.yaml: |
    %s
`, name, renderLabelsBlock(labels, 4), indentedConfig)
}

func renderDeploymentYAML(name, listenAddress, image, configMapName string, labels map[string]string) string {
	containerPort := listenAddressPort(listenAddress)
	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
%sspec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: %s
  template:
    metadata:
      labels:
%s    spec:
      volumes:
        - name: opa-config
          configMap:
            name: %s
      containers:
        - name: opa
          image: %s
          ports:
            - name: http
              containerPort: %d
              protocol: TCP
          readinessProbe:
            httpGet:
              path: /health?plugins
              port: %d
          livenessProbe:
            httpGet:
              path: /health
              port: %d
          volumeMounts:
            - name: opa-config
              mountPath: /config
              readOnly: true
          args:
            - run
            - --server
            - --addr=%s
            - --config-file=/config/opa-config.yaml
`, name, renderLabelsBlock(labels, 4), name, renderLabelsBlock(labels, 8), configMapName, image, containerPort, containerPort, containerPort, listenAddress)
}

func listenAddressPort(listenAddress string) int {
	trimmed := strings.TrimSpace(listenAddress)
	if trimmed == "" {
		return 8181
	}
	if strings.HasPrefix(trimmed, ":") {
		if port, err := strconv.Atoi(strings.TrimPrefix(trimmed, ":")); err == nil {
			return port
		}
	}
	host, portText, err := net.SplitHostPort(trimmed)
	if err != nil {
		return 8181
	}
	if host == "" && portText == "" {
		return 8181
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return 8181
	}
	return port
}

func builtInTopicLabels(appName, tenantName, topicName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":      deploymentName(appName, tenantName, topicName),
		"app.kubernetes.io/component": "opa",
		"app.kubernetes.io/tenant":    tenantName,
		"app.kubernetes.io/topic":     topicName,
	}
}

func mergeTopicLabels(builtIn, topicLabels map[string]string) map[string]string {
	merged := make(map[string]string, len(builtIn)+len(topicLabels))
	for key, value := range builtIn {
		merged[key] = value
	}
	for key, value := range topicLabels {
		if _, exists := merged[key]; exists {
			continue
		}
		merged[key] = value
	}
	return merged
}

func renderLabelsBlock(labels map[string]string, indent int) string {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	prefix := strings.Repeat(" ", indent)

	var b strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&b, "%s%s: %s\n", prefix, key, labels[key])
	}
	return b.String()
}

func deploymentName(appName, tenantName, topicName string) string {
	return fmt.Sprintf("%s-%s-%s-opa", sanitizeName(appName), sanitizeName(tenantName), sanitizeName(topicName))
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
