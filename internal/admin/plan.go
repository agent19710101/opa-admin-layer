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
	Name                       string            `json:"name"`
	BundleURL                  string            `json:"bundleURL"`
	DecisionPath               string            `json:"decisionPath"`
	ListenAddress              string            `json:"listenAddress"`
	Namespace                  string            `json:"namespace,omitempty"`
	Labels                     map[string]string `json:"labels,omitempty"`
	OPAConfigYAML              string            `json:"opaConfigYAML"`
	ConfigMapManifestYAML      string            `json:"configMapManifestYAML"`
	ServiceAccountManifestYAML string            `json:"serviceAccountManifestYAML,omitempty"`
	DeploymentManifestYAML     string            `json:"deploymentManifestYAML"`
	ServiceManifestYAML        string            `json:"serviceManifestYAML"`
	HPAManifestYAML            string            `json:"hpaManifestYAML,omitempty"`
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
	sharedServiceAccountBindings := countEffectiveServiceAccounts(normalized)
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
			effectiveConfigMapAnnotations := mergeStringMapWithRemovals(normalized.ControlPlane.ConfigMapAnnotations, topic.ConfigMapAnnotations, topic.RemoveConfigMapAnnotations)
			effectiveConfigMapLabels := mergeStringMapWithRemovals(normalized.ControlPlane.ConfigMapLabels, topic.ConfigMapLabels, topic.RemoveConfigMapLabels)
			renderedConfigMapLabels := mergeProtectedStringMap(renderedLabels, effectiveConfigMapLabels, builtInLabels)
			effectiveServiceType := normalized.ControlPlane.ServiceType
			if topic.ServiceType != "" {
				effectiveServiceType = topic.ServiceType
			}
			effectiveServiceAnnotations := mergeStringMapWithRemovals(normalized.ControlPlane.ServiceAnnotations, topic.ServiceAnnotations, topic.RemoveServiceAnnotations)
			effectiveServiceLabels := mergeStringMapWithRemovals(normalized.ControlPlane.ServiceLabels, topic.ServiceLabels, topic.RemoveServiceLabels)
			renderedServiceLabels := mergeProtectedStringMap(renderedLabels, effectiveServiceLabels, builtInLabels)
			effectiveDeploymentAnnotations := mergeStringMapWithRemovals(normalized.ControlPlane.DeploymentAnnotations, topic.DeploymentAnnotations, topic.RemoveDeploymentAnnotations)
			effectiveDeploymentLabels := mergeStringMapWithRemovals(normalized.ControlPlane.DeploymentLabels, topic.DeploymentLabels, topic.RemoveDeploymentLabels)
			renderedDeploymentLabels := mergeProtectedStringMap(renderedLabels, effectiveDeploymentLabels, builtInLabels)
			effectivePodAnnotations := mergeStringMapWithRemovals(normalized.ControlPlane.PodAnnotations, topic.PodAnnotations, topic.RemovePodAnnotations)
			effectivePodLabels := mergeStringMapWithRemovals(normalized.ControlPlane.PodLabels, topic.PodLabels, topic.RemovePodLabels)
			renderedPodLabels := mergeProtectedStringMap(renderedLabels, effectivePodLabels, builtInLabels)
			effectiveServiceAccountName := normalized.ControlPlane.ServiceAccountName
			if topic.ServiceAccountName != "" {
				effectiveServiceAccountName = topic.ServiceAccountName
			}
			effectiveServiceAccountAnnotations := mergeStringMapWithRemovals(normalized.ControlPlane.ServiceAccountAnnotations, topic.ServiceAccountAnnotations, topic.RemoveServiceAccountAnnotations)
			effectiveServiceAccountLabels := mergeStringMapWithRemovals(normalized.ControlPlane.ServiceAccountLabels, topic.ServiceAccountLabels, topic.RemoveServiceAccountLabels)
			renderedServiceAccountLabels := mergeProtectedStringMap(renderedLabels, effectiveServiceAccountLabels, builtInLabels)
			effectiveAutomountServiceAccountToken := normalized.ControlPlane.AutomountServiceAccountToken
			if topic.AutomountServiceAccountToken != nil {
				effectiveAutomountServiceAccountToken = topic.AutomountServiceAccountToken
			}
			effectiveExternalTrafficPolicy := normalized.ControlPlane.ExternalTrafficPolicy
			if topic.ExternalTrafficPolicy != "" {
				effectiveExternalTrafficPolicy = topic.ExternalTrafficPolicy
			}
			effectiveInternalTrafficPolicy := normalized.ControlPlane.InternalTrafficPolicy
			if topic.InternalTrafficPolicy != "" {
				effectiveInternalTrafficPolicy = topic.InternalTrafficPolicy
			}
			effectiveSessionAffinity := normalized.ControlPlane.SessionAffinity
			if topic.SessionAffinity != "" {
				effectiveSessionAffinity = topic.SessionAffinity
			}
			effectiveReplicas := normalized.ControlPlane.Replicas
			if topic.Replicas != 0 {
				effectiveReplicas = topic.Replicas
			}
			var effectiveAutoscaling *Autoscaling
			if normalized.ControlPlane.Autoscaling != nil {
				effectiveAutoscaling = normalized.ControlPlane.Autoscaling
			}
			if topic.Autoscaling != nil {
				effectiveAutoscaling = topic.Autoscaling
			}
			if effectiveAutoscaling != nil {
				effectiveReplicas = effectiveAutoscaling.MinReplicas
			}
			effectiveImagePullPolicy := normalized.ControlPlane.ImagePullPolicy
			if topic.ImagePullPolicy != "" {
				effectiveImagePullPolicy = topic.ImagePullPolicy
			}
			var hpaManifestYAML string
			if effectiveAutoscaling != nil {
				hpaManifestYAML = renderHPAYAML(workloadName, normalized.ControlPlane.Namespace, effectiveAutoscaling)
			}
			var serviceAccountManifestYAML string
			if effectiveServiceAccountName != "" && sharedServiceAccountBindings[strings.ToLower(effectiveServiceAccountName)] <= 1 {
				serviceAccountManifestYAML = renderServiceAccountYAML(effectiveServiceAccountName, normalized.ControlPlane.Namespace, renderedServiceAccountLabels, effectiveServiceAccountAnnotations)
			}
			tenantPlan.Topics = append(tenantPlan.Topics, TopicPlan{
				Name:                       topic.Name,
				BundleURL:                  bundleURL,
				DecisionPath:               topic.DecisionPath,
				ListenAddress:              normalized.ControlPlane.DefaultListenAddress,
				Namespace:                  normalized.ControlPlane.Namespace,
				Labels:                     topic.Labels,
				OPAConfigYAML:              opaConfigYAML,
				ConfigMapManifestYAML:      renderConfigMapYAML(configMapName, normalized.ControlPlane.Namespace, opaConfigYAML, renderedConfigMapLabels, effectiveConfigMapAnnotations),
				ServiceAccountManifestYAML: serviceAccountManifestYAML,
				DeploymentManifestYAML:     renderDeploymentYAML(workloadName, normalized.ControlPlane.Namespace, effectiveReplicas, normalized.ControlPlane.DefaultListenAddress, listenPort, normalized.ControlPlane.OPAImage, effectiveImagePullPolicy, configMapName, renderedDeploymentLabels, effectiveDeploymentAnnotations, effectivePodAnnotations, renderedPodLabels, effectiveServiceAccountName, effectiveAutomountServiceAccountToken, effectiveResources),
				ServiceManifestYAML:        renderServiceYAML(serviceName(normalized.Name, tenant.Name, topic.Name), normalized.ControlPlane.Namespace, workloadName, effectiveServiceType, effectiveExternalTrafficPolicy, effectiveInternalTrafficPolicy, effectiveSessionAffinity, listenPort, renderedServiceLabels, effectiveServiceAnnotations),
				HPAManifestYAML:            hpaManifestYAML,
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

func renderConfigMapYAML(name, namespace, opaConfig string, labels, annotations map[string]string) string {
	indentedConfig := strings.ReplaceAll(strings.TrimRight(opaConfig, "\n"), "\n", "\n    ")
	return fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
%s%s  labels:
%sdata:
  opa-config.yaml: |
    %s
`, name, renderNamespaceSection(namespace, 2), renderAnnotationsSection(annotations, 2), renderStringMapBlock(labels, 4), indentedConfig)
}

func renderServiceAccountYAML(name, namespace string, labels, annotations map[string]string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: %s
%s%s%s`, name, renderNamespaceSection(namespace, 2), renderAnnotationsSection(annotations, 2), renderLabelsSection(labels, 2))
}

func renderDeploymentYAML(name, namespace string, replicas int, listenAddress string, containerPort int, image, imagePullPolicy, configMapName string, deploymentLabels, deploymentAnnotations, podAnnotations, podLabels map[string]string, serviceAccountName string, automountServiceAccountToken *bool, resources ResourceRequirements) string {
	return fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
%s%s  labels:
%sspec:
  replicas: %d
  selector:
    matchLabels:
      app.kubernetes.io/name: %s
  template:
    metadata:
%s      labels:
%s    spec:
%s%s      volumes:
        - name: opa-config
          configMap:
            name: %s
      containers:
        - name: opa
          image: %s
%s          ports:
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
`, name, renderNamespaceSection(namespace, 2), renderAnnotationsSection(deploymentAnnotations, 2), renderStringMapBlock(deploymentLabels, 4), replicas, name, renderAnnotationsSection(podAnnotations, 6), renderStringMapBlock(podLabels, 8), renderServiceAccountNameSection(serviceAccountName, 6), renderAutomountServiceAccountTokenSection(automountServiceAccountToken, 6), configMapName, image, renderImagePullPolicySection(imagePullPolicy, 10), containerPort, containerPort, containerPort, renderResourcesBlock(resources, 10), listenAddress)
}

func renderServiceYAML(name, namespace, workloadName, serviceType, externalTrafficPolicy, internalTrafficPolicy, sessionAffinity string, port int, labels, annotations map[string]string) string {
	return fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
%s%s  labels:
%sspec:
  type: %s
%s%s%s  selector:
    app.kubernetes.io/name: %s
  ports:
    - name: http
      port: %d
      targetPort: %d
      protocol: TCP
`, name, renderNamespaceSection(namespace, 2), renderAnnotationsSection(annotations, 2), renderStringMapBlock(labels, 4), serviceType, renderExternalTrafficPolicySection(externalTrafficPolicy, 2), renderInternalTrafficPolicySection(internalTrafficPolicy, 2), renderSessionAffinitySection(sessionAffinity, 2), workloadName, port, port)
}

func renderHPAYAML(name, namespace string, autoscaling *Autoscaling) string {
	var metrics strings.Builder
	if autoscaling.TargetCPUUtilizationPercentage != 0 {
		fmt.Fprintf(&metrics, `    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: %d
`, autoscaling.TargetCPUUtilizationPercentage)
	}
	if autoscaling.TargetMemoryUtilizationPercentage != 0 {
		fmt.Fprintf(&metrics, `    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: %d
`, autoscaling.TargetMemoryUtilizationPercentage)
	}
	return fmt.Sprintf(`apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: %s
%sspec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: %s
  minReplicas: %d
  maxReplicas: %d
  metrics:
%s%s`, name, renderNamespaceSection(namespace, 2), name, autoscaling.MinReplicas, autoscaling.MaxReplicas, metrics.String(), renderHPABehaviorSection(autoscaling.Behavior, 2))
}

func renderHPABehaviorSection(value *AutoscalingBehavior, indent int) string {
	if value == nil {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%sbehavior:\n", strings.Repeat(" ", indent))
	if rendered := renderHPABehaviorPolicySection("scaleUp", value.ScaleUp, indent+2); rendered != "" {
		b.WriteString(rendered)
	}
	if rendered := renderHPABehaviorPolicySection("scaleDown", value.ScaleDown, indent+2); rendered != "" {
		b.WriteString(rendered)
	}
	return b.String()
}

func renderHPABehaviorPolicySection(name string, value *AutoscalingBehaviorPolicy, indent int) string {
	if value == nil {
		return ""
	}
	var b strings.Builder
	prefix := strings.Repeat(" ", indent)
	fmt.Fprintf(&b, "%s%s:\n", prefix, name)
	if value.StabilizationWindowSeconds != nil {
		fmt.Fprintf(&b, "%s  stabilizationWindowSeconds: %d\n", prefix, *value.StabilizationWindowSeconds)
	}
	if selectPolicy := strings.TrimSpace(value.SelectPolicy); selectPolicy != "" {
		fmt.Fprintf(&b, "%s  selectPolicy: %s\n", prefix, selectPolicy)
	}
	if len(value.Policies) > 0 {
		fmt.Fprintf(&b, "%s  policies:\n", prefix)
		for _, policy := range value.Policies {
			fmt.Fprintf(&b, "%s    - type: %s\n", prefix, strings.TrimSpace(policy.Type))
			fmt.Fprintf(&b, "%s      value: %d\n", prefix, policy.Value)
			fmt.Fprintf(&b, "%s      periodSeconds: %d\n", prefix, policy.PeriodSeconds)
		}
	}
	return b.String()
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

func countEffectiveServiceAccounts(spec Specification) map[string]int {
	counts := map[string]int{}
	sharedName := strings.TrimSpace(spec.ControlPlane.ServiceAccountName)
	for _, tenant := range spec.Tenants {
		for _, topic := range tenant.Topics {
			effectiveName := sharedName
			if topicName := strings.TrimSpace(topic.ServiceAccountName); topicName != "" {
				effectiveName = topicName
			}
			if effectiveName == "" {
				continue
			}
			counts[strings.ToLower(effectiveName)]++
		}
	}
	return counts
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

func mergeStringMapWithRemovals(base, override map[string]string, removals []string) map[string]string {
	merged := mergeStringMap(base, override)
	if len(merged) == 0 && len(removals) == 0 {
		return nil
	}
	for _, key := range removals {
		delete(merged, strings.TrimSpace(key))
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}

func mergeProtectedStringMap(base, override, protected map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	merged := make(map[string]string, len(base)+len(override))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range override {
		if _, blocked := protected[key]; blocked {
			continue
		}
		merged[key] = value
	}
	return merged
}

func renderNamespaceSection(namespace string, indent int) string {
	trimmed := strings.TrimSpace(namespace)
	if trimmed == "" {
		return ""
	}
	return fmt.Sprintf("%snamespace: %s\n", strings.Repeat(" ", indent), trimmed)
}

func renderServiceAccountNameSection(value string, indent int) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return fmt.Sprintf("%sserviceAccountName: %s\n", strings.Repeat(" ", indent), trimmed)
}

func renderImagePullPolicySection(value string, indent int) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return fmt.Sprintf("%simagePullPolicy: %s\n", strings.Repeat(" ", indent), trimmed)
}

func renderAutomountServiceAccountTokenSection(value *bool, indent int) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%sautomountServiceAccountToken: %t\n", strings.Repeat(" ", indent), *value)
}

func renderExternalTrafficPolicySection(policy string, indent int) string {
	trimmed := strings.TrimSpace(policy)
	if trimmed == "" {
		return ""
	}
	return fmt.Sprintf("%sexternalTrafficPolicy: %s\n", strings.Repeat(" ", indent), trimmed)
}

func renderInternalTrafficPolicySection(policy string, indent int) string {
	trimmed := strings.TrimSpace(policy)
	if trimmed == "" {
		return ""
	}
	return fmt.Sprintf("%sinternalTrafficPolicy: %s\n", strings.Repeat(" ", indent), trimmed)
}

func renderSessionAffinitySection(value string, indent int) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return fmt.Sprintf("%ssessionAffinity: %s\n", strings.Repeat(" ", indent), trimmed)
}

func renderAnnotationsSection(annotations map[string]string, indent int) string {
	if len(annotations) == 0 {
		return ""
	}
	return fmt.Sprintf("%sannotations:\n%s", strings.Repeat(" ", indent), renderStringMapBlock(annotations, indent+2))
}

func renderLabelsSection(labels map[string]string, indent int) string {
	if len(labels) == 0 {
		return ""
	}
	return fmt.Sprintf("%slabels:\n%s", strings.Repeat(" ", indent), renderStringMapBlock(labels, indent+2))
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
