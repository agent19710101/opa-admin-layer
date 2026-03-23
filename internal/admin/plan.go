package admin

import (
	"encoding/json"
	"errors"
	"fmt"
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
	ServiceManifestYAML    string            `json:"serviceManifestYAML"`
}

func BuildPlan(spec Specification) (Plan, error) {
	if issues := Validate(spec); len(issues) > 0 {
		return Plan{}, errors.New(strings.Join(issues, "; "))
	}
	normalized := normalize(spec)
	listenPort, err := parseListenAddressPort(normalized.ControlPlane.DefaultListenAddress)
	if err != nil {
		return Plan{}, fmt.Errorf("normalized controlPlane.defaultListenAddress is invalid: %w", err)
	}
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
			workloadName := deploymentName(normalized.Name, tenant.Name, topic.Name)
			builtInLabels := builtInTopicLabels(normalized.Name, tenant.Name, topic.Name)
			renderedLabels := mergeTopicLabels(builtInLabels, topic.Labels)
			configMapName := topicConfigMapName(normalized.Name, tenant.Name, topic.Name)
			effectiveResources := mergeResourceRequirements(normalized.ControlPlane.OPAResources, topic.OPAResources)
			effectiveServiceType := normalized.ControlPlane.ServiceType
			if topic.ServiceType != "" {
				effectiveServiceType = topic.ServiceType
			}
			effectiveServiceAnnotations := mergeStringMap(normalized.ControlPlane.ServiceAnnotations, topic.ServiceAnnotations)
			effectiveExternalTrafficPolicy := normalized.ControlPlane.ExternalTrafficPolicy
			if topic.ExternalTrafficPolicy != "" {
				effectiveExternalTrafficPolicy = topic.ExternalTrafficPolicy
			}
			tenantPlan.Topics = append(tenantPlan.Topics, TopicPlan{
				Name:                   topic.Name,
				BundleURL:              bundleURL,
				DecisionPath:           topic.DecisionPath,
				ListenAddress:          normalized.ControlPlane.DefaultListenAddress,
				Labels:                 topic.Labels,
				OPAConfigYAML:          opaConfigYAML,
				ConfigMapManifestYAML:  renderConfigMapYAML(configMapName, opaConfigYAML, renderedLabels),
				DeploymentManifestYAML: renderDeploymentYAML(workloadName, normalized.ControlPlane.DefaultListenAddress, listenPort, normalized.ControlPlane.OPAImage, configMapName, renderedLabels, effectiveResources),
				ServiceManifestYAML:    renderServiceYAML(serviceName(normalized.Name, tenant.Name, topic.Name), workloadName, effectiveServiceType, effectiveExternalTrafficPolicy, listenPort, renderedLabels, effectiveServiceAnnotations),
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
`, name, renderStringMapBlock(labels, 4), indentedConfig)
}

func renderDeploymentYAML(name, listenAddress string, containerPort int, image, configMapName string, labels map[string]string, resources ResourceRequirements) string {
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
%s          volumeMounts:
            - name: opa-config
              mountPath: /config
              readOnly: true
          args:
            - run
            - --server
            - --addr=%s
            - --config-file=/config/opa-config.yaml
`, name, renderStringMapBlock(labels, 4), name, renderStringMapBlock(labels, 8), configMapName, image, containerPort, containerPort, containerPort, renderResourcesBlock(resources, 10), listenAddress)
}

func renderServiceYAML(name, workloadName, serviceType, externalTrafficPolicy string, port int, labels, annotations map[string]string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
%s  labels:
%sspec:
  type: %s
%s  selector:
    app.kubernetes.io/name: %s
  ports:
    - name: http
      port: %d
      targetPort: %d
      protocol: TCP
`, name, renderAnnotationsSection(annotations, 2), renderStringMapBlock(labels, 4), serviceType, renderExternalTrafficPolicySection(externalTrafficPolicy, 2), workloadName, port, port)
}

func renderResourcesBlock(resources ResourceRequirements, indent int) string {
	var b strings.Builder
	prefix := strings.Repeat(" ", indent)
	if rendered := renderResourceListBlock("requests", resources.Requests, indent+2); rendered != "" {
		if b.Len() == 0 {
			fmt.Fprintf(&b, "%sresources:\n", prefix)
		}
		b.WriteString(rendered)
	}
	if rendered := renderResourceListBlock("limits", resources.Limits, indent+2); rendered != "" {
		if b.Len() == 0 {
			fmt.Fprintf(&b, "%sresources:\n", prefix)
		}
		b.WriteString(rendered)
	}
	return b.String()
}

func renderResourceListBlock(name string, values *ResourceList, indent int) string {
	if values == nil {
		return ""
	}
	var b strings.Builder
	prefix := strings.Repeat(" ", indent)
	if cpu := strings.TrimSpace(values.CPU); cpu != "" {
		fmt.Fprintf(&b, "%s%s:\n", prefix, name)
		fmt.Fprintf(&b, "%s  cpu: %s\n", prefix, strconv.Quote(cpu))
		if memory := strings.TrimSpace(values.Memory); memory != "" {
			fmt.Fprintf(&b, "%s  memory: %s\n", prefix, strconv.Quote(memory))
		}
		return b.String()
	}
	if memory := strings.TrimSpace(values.Memory); memory != "" {
		fmt.Fprintf(&b, "%s%s:\n", prefix, name)
		fmt.Fprintf(&b, "%s  memory: %s\n", prefix, strconv.Quote(memory))
	}
	return b.String()
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

func renderExternalTrafficPolicySection(policy string, indent int) string {
	trimmed := strings.TrimSpace(policy)
	if trimmed == "" {
		return ""
	}
	return fmt.Sprintf("%sexternalTrafficPolicy: %s\n", strings.Repeat(" ", indent), trimmed)
}

func renderAnnotationsSection(annotations map[string]string, indent int) string {
	if len(annotations) == 0 {
		return ""
	}
	return fmt.Sprintf("%sannotations:\n%s", strings.Repeat(" ", indent), renderStringMapBlock(annotations, indent+2))
}

func renderStringMapBlock(values map[string]string, indent int) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	prefix := strings.Repeat(" ", indent)

	var b strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&b, "%s%s: %s\n", prefix, key, strconv.Quote(values[key]))
	}
	return b.String()
}

func deploymentName(appName, tenantName, topicName string) string {
	return fmt.Sprintf("%s-%s-%s-opa", sanitizeName(appName), sanitizeName(tenantName), sanitizeName(topicName))
}

func topicConfigMapName(appName, tenantName, topicName string) string {
	return fmt.Sprintf("%s-%s-%s-opa-config", sanitizeName(appName), sanitizeName(tenantName), sanitizeName(topicName))
}

func serviceName(appName, tenantName, topicName string) string {
	return fmt.Sprintf("%s-%s-%s-opa", sanitizeName(appName), sanitizeName(tenantName), sanitizeName(topicName))
}

func sanitizeName(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer("_", "-", " ", "-")
	return replacer.Replace(value)
}

func MarshalPlan(plan Plan) ([]byte, error) {
	return json.MarshalIndent(plan, "", "  ")
}
