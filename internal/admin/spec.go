package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/resource"
)

const DefaultOPAImage = "openpolicyagent/opa:1.12.1"

var (
	kubernetesLabelNamePattern  = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9_.-]{0,61}[A-Za-z0-9])?$`)
	kubernetesLabelValuePattern = regexp.MustCompile(`^([A-Za-z0-9]([A-Za-z0-9_.-]{0,61}[A-Za-z0-9])?)?$`)
	dns1123LabelPattern         = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
)

type Specification struct {
	Name         string       `json:"name" yaml:"name"`
	ControlPlane ControlPlane `json:"controlPlane" yaml:"controlPlane"`
	Tenants      []Tenant     `json:"tenants" yaml:"tenants"`
}

type ControlPlane struct {
	BaseServiceURL        string               `json:"baseServiceURL" yaml:"baseServiceURL"`
	BundlePrefix          string               `json:"bundlePrefix" yaml:"bundlePrefix"`
	DefaultDecisionPath   string               `json:"defaultDecisionPath" yaml:"defaultDecisionPath"`
	DefaultListenAddress  string               `json:"defaultListenAddress" yaml:"defaultListenAddress"`
	OPAImage              string               `json:"opaImage" yaml:"opaImage"`
	Namespace             string               `json:"namespace" yaml:"namespace"`
	Replicas              int                  `json:"replicas" yaml:"replicas"`
	ServiceType           string               `json:"serviceType" yaml:"serviceType"`
	ServiceAnnotations    map[string]string    `json:"serviceAnnotations" yaml:"serviceAnnotations"`
	DeploymentAnnotations map[string]string    `json:"deploymentAnnotations" yaml:"deploymentAnnotations"`
	PodAnnotations        map[string]string    `json:"podAnnotations" yaml:"podAnnotations"`
	ExternalTrafficPolicy string               `json:"externalTrafficPolicy" yaml:"externalTrafficPolicy"`
	InternalTrafficPolicy string               `json:"internalTrafficPolicy" yaml:"internalTrafficPolicy"`
	SessionAffinity       string               `json:"sessionAffinity" yaml:"sessionAffinity"`
	OPAResources          ResourceRequirements `json:"opaResources" yaml:"opaResources"`
}

type ResourceRequirements struct {
	Requests *ResourceList `json:"requests,omitempty" yaml:"requests,omitempty"`
	Limits   *ResourceList `json:"limits,omitempty" yaml:"limits,omitempty"`
}

type ResourceList struct {
	CPU    string `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	Memory string `json:"memory,omitempty" yaml:"memory,omitempty"`
}

type Tenant struct {
	Name   string  `json:"name" yaml:"name"`
	Topics []Topic `json:"topics" yaml:"topics"`
}

type Topic struct {
	Name                  string               `json:"name" yaml:"name"`
	BundleResource        string               `json:"bundleResource,omitempty" yaml:"bundleResource,omitempty"`
	DecisionPath          string               `json:"decisionPath,omitempty" yaml:"decisionPath,omitempty"`
	Labels                map[string]string    `json:"labels,omitempty" yaml:"labels,omitempty"`
	Replicas              int                  `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	ServiceType           string               `json:"serviceType,omitempty" yaml:"serviceType,omitempty"`
	ServiceAnnotations    map[string]string    `json:"serviceAnnotations,omitempty" yaml:"serviceAnnotations,omitempty"`
	DeploymentAnnotations map[string]string    `json:"deploymentAnnotations,omitempty" yaml:"deploymentAnnotations,omitempty"`
	PodAnnotations        map[string]string    `json:"podAnnotations,omitempty" yaml:"podAnnotations,omitempty"`
	ExternalTrafficPolicy string               `json:"externalTrafficPolicy,omitempty" yaml:"externalTrafficPolicy,omitempty"`
	InternalTrafficPolicy string               `json:"internalTrafficPolicy,omitempty" yaml:"internalTrafficPolicy,omitempty"`
	SessionAffinity       string               `json:"sessionAffinity,omitempty" yaml:"sessionAffinity,omitempty"`
	OPAResources          ResourceRequirements `json:"opaResources,omitempty" yaml:"opaResources,omitempty"`
}

func LoadSpec(path string) (Specification, error) {
	payload, err := os.ReadFile(path)
	if err != nil {
		return Specification{}, fmt.Errorf("read spec: %w", err)
	}
	spec, err := DecodeSpec(payload)
	if err != nil {
		return Specification{}, fmt.Errorf("decode spec: %w", err)
	}
	return spec, nil
}

// DecodeSpec decodes a specification from JSON or YAML and rejects unknown fields.
func DecodeSpec(payload []byte) (Specification, error) {
	if looksLikeJSON(payload) {
		return decodeJSONSpec(payload)
	}
	return decodeYAMLSpec(payload)
}

func decodeJSONSpec(payload []byte) (Specification, error) {
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()

	var spec Specification
	if err := dec.Decode(&spec); err != nil {
		return Specification{}, err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Specification{}, fmt.Errorf("spec must contain exactly one JSON object")
		}
		return Specification{}, err
	}
	return spec, nil
}

func decodeYAMLSpec(payload []byte) (Specification, error) {
	dec := yaml.NewDecoder(bytes.NewReader(payload))
	dec.KnownFields(true)

	var spec Specification
	if err := dec.Decode(&spec); err != nil {
		return Specification{}, err
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil && extra != nil {
			return Specification{}, fmt.Errorf("spec must contain exactly one YAML document")
		}
		return Specification{}, err
	}
	return spec, nil
}

func looksLikeJSON(payload []byte) bool {
	trimmed := bytes.TrimSpace(payload)
	return len(trimmed) > 0 && trimmed[0] == '{'
}

func Validate(spec Specification) []string {
	var issues []string
	if strings.TrimSpace(spec.Name) == "" {
		issues = append(issues, "spec.name must not be empty")
	}
	if err := validateBaseServiceURL(spec.ControlPlane.BaseServiceURL); err != nil {
		issues = append(issues, fmt.Sprintf("controlPlane.baseServiceURL is invalid: %v", err))
	}
	if err := validateListenAddress(spec.ControlPlane.DefaultListenAddress); err != nil {
		issues = append(issues, fmt.Sprintf("controlPlane.defaultListenAddress is invalid: %v", err))
	}
	if err := validateNamespace(spec.ControlPlane.Namespace); err != nil {
		issues = append(issues, fmt.Sprintf("controlPlane.namespace is invalid: %v", err))
	}
	if err := validateReplicas(spec.ControlPlane.Replicas); err != nil {
		issues = append(issues, fmt.Sprintf("controlPlane.replicas is invalid: %v", err))
	}
	seenTenants := map[string]struct{}{}
	if err := validateKubernetesServiceType(spec.ControlPlane.ServiceType); err != nil {
		issues = append(issues, fmt.Sprintf("controlPlane.serviceType is invalid: %v", err))
	}
	for annotationKey := range spec.ControlPlane.ServiceAnnotations {
		if err := validateKubernetesLabelKey(annotationKey); err != nil {
			issues = append(issues, fmt.Sprintf("controlPlane.serviceAnnotations key %q is invalid: %v", annotationKey, err))
		}
	}
	for annotationKey := range spec.ControlPlane.DeploymentAnnotations {
		if err := validateKubernetesLabelKey(annotationKey); err != nil {
			issues = append(issues, fmt.Sprintf("controlPlane.deploymentAnnotations key %q is invalid: %v", annotationKey, err))
		}
	}
	for annotationKey := range spec.ControlPlane.PodAnnotations {
		if err := validateKubernetesLabelKey(annotationKey); err != nil {
			issues = append(issues, fmt.Sprintf("controlPlane.podAnnotations key %q is invalid: %v", annotationKey, err))
		}
	}
	if err := validateExternalTrafficPolicy(spec.ControlPlane.ExternalTrafficPolicy); err != nil {
		issues = append(issues, fmt.Sprintf("controlPlane.externalTrafficPolicy is invalid: %v", err))
	}
	if err := validateInternalTrafficPolicy(spec.ControlPlane.InternalTrafficPolicy); err != nil {
		issues = append(issues, fmt.Sprintf("controlPlane.internalTrafficPolicy is invalid: %v", err))
	}
	if err := validateSessionAffinity(spec.ControlPlane.SessionAffinity); err != nil {
		issues = append(issues, fmt.Sprintf("controlPlane.sessionAffinity is invalid: %v", err))
	}
	if err := validateServiceTrafficPolicyCompatibility(spec.ControlPlane.ServiceType, spec.ControlPlane.ExternalTrafficPolicy); err != nil {
		issues = append(issues, fmt.Sprintf("controlPlane.externalTrafficPolicy is invalid: %v", err))
	}
	issues = append(issues, validateOPAResources(spec.ControlPlane.OPAResources)...)
	issues = append(issues, validateOPAResourceBudgetAtPath("controlPlane.opaResources", normalizeResourceRequirements(spec.ControlPlane.OPAResources))...)
	if len(spec.Tenants) == 0 {
		issues = append(issues, "spec.tenants must contain at least one tenant")
	}
	for _, tenant := range spec.Tenants {
		tenantName := strings.TrimSpace(tenant.Name)
		if tenantName == "" {
			issues = append(issues, "tenant.name must not be empty")
			continue
		}
		key := strings.ToLower(tenantName)
		if _, ok := seenTenants[key]; ok {
			issues = append(issues, fmt.Sprintf("tenant.name %q is duplicated", tenantName))
		} else {
			seenTenants[key] = struct{}{}
		}
		if len(tenant.Topics) == 0 {
			issues = append(issues, fmt.Sprintf("tenant %q must define at least one topic", tenantName))
		}
		seenTopics := map[string]struct{}{}
		for _, topic := range tenant.Topics {
			topicName := strings.TrimSpace(topic.Name)
			if topicName == "" {
				issues = append(issues, fmt.Sprintf("tenant %q has empty topic name", tenantName))
				continue
			}
			topicKey := strings.ToLower(topicName)
			if _, ok := seenTopics[topicKey]; ok {
				issues = append(issues, fmt.Sprintf("tenant %q repeats topic %q", tenantName, topicName))
			} else {
				seenTopics[topicKey] = struct{}{}
			}
			if err := validateReplicas(topic.Replicas); err != nil {
				issues = append(issues, fmt.Sprintf("tenant %q topic %q replicas is invalid: %v", tenantName, topicName, err))
			}
			for labelKey, labelValue := range topic.Labels {
				if err := validateKubernetesLabelKey(labelKey); err != nil {
					issues = append(issues, fmt.Sprintf("tenant %q topic %q label key %q is invalid: %v", tenantName, topicName, labelKey, err))
				}
				if err := validateKubernetesLabelValue(labelValue); err != nil {
					issues = append(issues, fmt.Sprintf("tenant %q topic %q label %q has invalid value %q: %v", tenantName, topicName, labelKey, labelValue, err))
				}
			}
			if err := validateKubernetesServiceType(topic.ServiceType); err != nil {
				issues = append(issues, fmt.Sprintf("tenant %q topic %q serviceType is invalid: %v", tenantName, topicName, err))
			}
			for annotationKey := range topic.ServiceAnnotations {
				if err := validateKubernetesLabelKey(annotationKey); err != nil {
					issues = append(issues, fmt.Sprintf("tenant %q topic %q serviceAnnotations key %q is invalid: %v", tenantName, topicName, annotationKey, err))
				}
			}
			for annotationKey := range topic.DeploymentAnnotations {
				if err := validateKubernetesLabelKey(annotationKey); err != nil {
					issues = append(issues, fmt.Sprintf("tenant %q topic %q deploymentAnnotations key %q is invalid: %v", tenantName, topicName, annotationKey, err))
				}
			}
			for annotationKey := range topic.PodAnnotations {
				if err := validateKubernetesLabelKey(annotationKey); err != nil {
					issues = append(issues, fmt.Sprintf("tenant %q topic %q podAnnotations key %q is invalid: %v", tenantName, topicName, annotationKey, err))
				}
			}
			if err := validateExternalTrafficPolicy(topic.ExternalTrafficPolicy); err != nil {
				issues = append(issues, fmt.Sprintf("tenant %q topic %q externalTrafficPolicy is invalid: %v", tenantName, topicName, err))
			}
			if err := validateInternalTrafficPolicy(topic.InternalTrafficPolicy); err != nil {
				issues = append(issues, fmt.Sprintf("tenant %q topic %q internalTrafficPolicy is invalid: %v", tenantName, topicName, err))
			}
			if err := validateSessionAffinity(topic.SessionAffinity); err != nil {
				issues = append(issues, fmt.Sprintf("tenant %q topic %q sessionAffinity is invalid: %v", tenantName, topicName, err))
			}
			effectiveServiceType := strings.TrimSpace(spec.ControlPlane.ServiceType)
			if strings.TrimSpace(topic.ServiceType) != "" {
				effectiveServiceType = topic.ServiceType
			}
			effectiveExternalTrafficPolicy := strings.TrimSpace(spec.ControlPlane.ExternalTrafficPolicy)
			if strings.TrimSpace(topic.ExternalTrafficPolicy) != "" {
				effectiveExternalTrafficPolicy = topic.ExternalTrafficPolicy
			}
			if err := validateServiceTrafficPolicyCompatibility(effectiveServiceType, effectiveExternalTrafficPolicy); err != nil {
				issues = append(issues, fmt.Sprintf("tenant %q topic %q effective externalTrafficPolicy is invalid: %v", tenantName, topicName, err))
			}
			issues = append(issues, validateOPAResourcesAtPath(fmt.Sprintf("tenant %q topic %q opaResources", tenantName, topicName), topic.OPAResources)...)
			effectiveResources := mergeResourceRequirements(normalizeResourceRequirements(spec.ControlPlane.OPAResources), normalizeResourceRequirements(topic.OPAResources))
			issues = append(issues, validateOPAResourceBudgetAtPath(fmt.Sprintf("tenant %q topic %q effective opaResources", tenantName, topicName), effectiveResources)...)
			for resourceKind, resourceName := range map[string]string{
				"deployment": deploymentName(spec.Name, tenantName, topicName),
				"configmap":  topicConfigMapName(spec.Name, tenantName, topicName),
				"service":    serviceName(spec.Name, tenantName, topicName),
			} {
				if err := validateRenderedResourceName(resourceName); err != nil {
					issues = append(issues, fmt.Sprintf("tenant %q topic %q renders invalid %s name %q: %v", tenantName, topicName, resourceKind, resourceName, err))
				}
			}
		}
	}
	sort.Strings(issues)
	return issues
}

func validateKubernetesLabelKey(key string) error {
	parts := strings.Split(key, "/")
	switch len(parts) {
	case 1:
		return validateKubernetesLabelName(parts[0])
	case 2:
		if err := validateDNS1123Subdomain(parts[0]); err != nil {
			return fmt.Errorf("prefix must be a valid DNS subdomain: %w", err)
		}
		if err := validateKubernetesLabelName(parts[1]); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("must contain at most one '/' separator")
	}
}

func validateKubernetesLabelName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name must not be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("name must be 63 characters or fewer")
	}
	if !kubernetesLabelNamePattern.MatchString(name) {
		return fmt.Errorf("name must start and end with an alphanumeric character and contain only alphanumerics, '-', '_', or '.'")
	}
	return nil
}

func validateKubernetesLabelValue(value string) error {
	if len(value) > 63 {
		return fmt.Errorf("value must be 63 characters or fewer")
	}
	if !kubernetesLabelValuePattern.MatchString(value) {
		return fmt.Errorf("value must be empty or start and end with an alphanumeric character and contain only alphanumerics, '-', '_', or '.'")
	}
	return nil
}

func validateDNS1123Subdomain(value string) error {
	if len(value) == 0 {
		return fmt.Errorf("must not be empty")
	}
	if len(value) > 253 {
		return fmt.Errorf("must be 253 characters or fewer")
	}
	for _, label := range strings.Split(value, ".") {
		if !dns1123LabelPattern.MatchString(label) {
			return fmt.Errorf("segment %q must match DNS-1123 label syntax", label)
		}
	}
	return nil
}

func validateNamespace(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return validateRenderedResourceName(trimmed)
}

func validateReplicas(value int) error {
	if value < 0 {
		return fmt.Errorf("must be zero or greater")
	}
	return nil
}

func validateBaseServiceURL(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("must not be empty")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("must be a valid URL: %w", err)
	}
	if !parsed.IsAbs() {
		return fmt.Errorf("must be an absolute URL")
	}
	switch parsed.Scheme {
	case "http", "https":
		// ok
	default:
		return fmt.Errorf("scheme must be http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("host must not be empty")
	}
	if parsed.Fragment != "" {
		return fmt.Errorf("fragment is not supported")
	}
	return nil
}

func validateKubernetesServiceType(serviceType string) error {
	trimmed := strings.TrimSpace(serviceType)
	if trimmed == "" {
		return nil
	}
	switch trimmed {
	case "ClusterIP", "NodePort", "LoadBalancer":
		return nil
	default:
		return fmt.Errorf("must be one of ClusterIP, NodePort, or LoadBalancer")
	}
}

func validateExternalTrafficPolicy(policy string) error {
	trimmed := strings.TrimSpace(policy)
	if trimmed == "" {
		return nil
	}
	switch trimmed {
	case "Cluster", "Local":
		return nil
	default:
		return fmt.Errorf("must be Cluster or Local")
	}
}

func validateInternalTrafficPolicy(policy string) error {
	trimmed := strings.TrimSpace(policy)
	if trimmed == "" {
		return nil
	}
	switch trimmed {
	case "Cluster", "Local":
		return nil
	default:
		return fmt.Errorf("must be Cluster or Local")
	}
}

func validateSessionAffinity(value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	switch trimmed {
	case "None", "ClientIP":
		return nil
	default:
		return fmt.Errorf("must be None or ClientIP")
	}
}

func validateServiceTrafficPolicyCompatibility(serviceType, externalTrafficPolicy string) error {
	trimmedPolicy := strings.TrimSpace(externalTrafficPolicy)
	if trimmedPolicy == "" {
		return nil
	}
	switch strings.TrimSpace(serviceType) {
	case "NodePort", "LoadBalancer":
		return nil
	case "":
		return fmt.Errorf("requires serviceType NodePort or LoadBalancer, got ClusterIP")
	default:
		return fmt.Errorf("requires serviceType NodePort or LoadBalancer, got %s", strings.TrimSpace(serviceType))
	}
}

func validateListenAddress(listenAddress string) error {
	_, err := parseListenAddressPort(listenAddress)
	return err
}

func parseListenAddressPort(listenAddress string) (int, error) {
	trimmed := strings.TrimSpace(listenAddress)
	if trimmed == "" {
		return 8181, nil
	}

	_, portText, err := net.SplitHostPort(trimmed)
	if err != nil {
		return 0, fmt.Errorf("must use :port, host:port, or [ipv6]:port syntax")
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		return 0, fmt.Errorf("port must be numeric")
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port must be between 1 and 65535")
	}
	return port, nil
}

func validateRenderedResourceName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("must not be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("must be 63 characters or fewer")
	}
	if !dns1123LabelPattern.MatchString(name) {
		return fmt.Errorf("must use lowercase alphanumerics or '-', and start/end with an alphanumeric character")
	}
	return nil
}

func validateOPAResources(resources ResourceRequirements) []string {
	return validateOPAResourcesAtPath("controlPlane.opaResources", resources)
}

func validateOPAResourcesAtPath(path string, resources ResourceRequirements) []string {
	var issues []string
	issues = append(issues, validateResourceList(path+".requests", resources.Requests)...)
	issues = append(issues, validateResourceList(path+".limits", resources.Limits)...)
	return issues
}

func validateResourceList(path string, values *ResourceList) []string {
	if values == nil {
		return nil
	}
	var issues []string
	if strings.TrimSpace(values.CPU) == "" && strings.TrimSpace(values.Memory) == "" {
		issues = append(issues, fmt.Sprintf("%s must set cpu and/or memory", path))
		return issues
	}
	if cpu := strings.TrimSpace(values.CPU); cpu != "" {
		if err := validateKubernetesQuantity(cpu); err != nil {
			issues = append(issues, fmt.Sprintf("%s.cpu is invalid: %v", path, err))
		}
	}
	if memory := strings.TrimSpace(values.Memory); memory != "" {
		if err := validateKubernetesQuantity(memory); err != nil {
			issues = append(issues, fmt.Sprintf("%s.memory is invalid: %v", path, err))
		}
	}
	return issues
}

func validateKubernetesQuantity(value string) error {
	if _, err := resource.ParseQuantity(value); err != nil {
		return err
	}
	return nil
}

func validateOPAResourceBudgetAtPath(path string, resources ResourceRequirements) []string {
	var issues []string
	issues = append(issues, validateResourceBudget(path+".cpu", resources.Requests, resources.Limits, func(values *ResourceList) string {
		if values == nil {
			return ""
		}
		return values.CPU
	})...)
	issues = append(issues, validateResourceBudget(path+".memory", resources.Requests, resources.Limits, func(values *ResourceList) string {
		if values == nil {
			return ""
		}
		return values.Memory
	})...)
	return issues
}

func validateResourceBudget(path string, requests, limits *ResourceList, pick func(*ResourceList) string) []string {
	requestValue := strings.TrimSpace(pick(requests))
	limitValue := strings.TrimSpace(pick(limits))
	if requestValue == "" || limitValue == "" {
		return nil
	}
	requestQuantity, err := resource.ParseQuantity(requestValue)
	if err != nil {
		return nil
	}
	limitQuantity, err := resource.ParseQuantity(limitValue)
	if err != nil {
		return nil
	}
	if requestQuantity.Cmp(limitQuantity) > 0 {
		return []string{fmt.Sprintf("%s request %q must not exceed limit %q", path, requestValue, limitValue)}
	}
	return nil
}

func normalize(spec Specification) Specification {
	normalized := spec
	if strings.TrimSpace(normalized.ControlPlane.BundlePrefix) == "" {
		normalized.ControlPlane.BundlePrefix = "bundles"
	}
	if strings.TrimSpace(normalized.ControlPlane.DefaultDecisionPath) == "" {
		normalized.ControlPlane.DefaultDecisionPath = "system/authz/allow"
	}
	if strings.TrimSpace(normalized.ControlPlane.DefaultListenAddress) == "" {
		normalized.ControlPlane.DefaultListenAddress = ":8181"
	}
	if strings.TrimSpace(normalized.ControlPlane.OPAImage) == "" {
		normalized.ControlPlane.OPAImage = DefaultOPAImage
	}
	normalized.ControlPlane.Namespace = strings.TrimSpace(normalized.ControlPlane.Namespace)
	if normalized.ControlPlane.Replicas == 0 {
		normalized.ControlPlane.Replicas = 1
	}
	normalized.ControlPlane.ServiceType = strings.TrimSpace(normalized.ControlPlane.ServiceType)
	if normalized.ControlPlane.ServiceType == "" {
		normalized.ControlPlane.ServiceType = "ClusterIP"
	}
	normalized.ControlPlane.ExternalTrafficPolicy = strings.TrimSpace(normalized.ControlPlane.ExternalTrafficPolicy)
	normalized.ControlPlane.InternalTrafficPolicy = strings.TrimSpace(normalized.ControlPlane.InternalTrafficPolicy)
	normalized.ControlPlane.SessionAffinity = strings.TrimSpace(normalized.ControlPlane.SessionAffinity)
	normalized.ControlPlane.OPAResources = normalizeResourceRequirements(normalized.ControlPlane.OPAResources)
	for i := range normalized.Tenants {
		normalized.Tenants[i].Name = strings.TrimSpace(normalized.Tenants[i].Name)
		for j := range normalized.Tenants[i].Topics {
			normalized.Tenants[i].Topics[j].Name = strings.TrimSpace(normalized.Tenants[i].Topics[j].Name)
			if normalized.Tenants[i].Topics[j].Replicas == 0 {
				normalized.Tenants[i].Topics[j].Replicas = normalized.ControlPlane.Replicas
			}
			normalized.Tenants[i].Topics[j].ServiceType = strings.TrimSpace(normalized.Tenants[i].Topics[j].ServiceType)
			normalized.Tenants[i].Topics[j].ExternalTrafficPolicy = strings.TrimSpace(normalized.Tenants[i].Topics[j].ExternalTrafficPolicy)
			normalized.Tenants[i].Topics[j].InternalTrafficPolicy = strings.TrimSpace(normalized.Tenants[i].Topics[j].InternalTrafficPolicy)
			normalized.Tenants[i].Topics[j].SessionAffinity = strings.TrimSpace(normalized.Tenants[i].Topics[j].SessionAffinity)
			if strings.TrimSpace(normalized.Tenants[i].Topics[j].DecisionPath) == "" {
				normalized.Tenants[i].Topics[j].DecisionPath = normalized.ControlPlane.DefaultDecisionPath
			}
			if strings.TrimSpace(normalized.Tenants[i].Topics[j].BundleResource) == "" {
				normalized.Tenants[i].Topics[j].BundleResource = fmt.Sprintf("%s/%s/%s.tar.gz", strings.Trim(normalized.ControlPlane.BundlePrefix, "/"), normalized.Tenants[i].Name, normalized.Tenants[i].Topics[j].Name)
			}
			normalized.Tenants[i].Topics[j].OPAResources = normalizeResourceRequirements(normalized.Tenants[i].Topics[j].OPAResources)
		}
	}
	return normalized
}

func normalizeResourceRequirements(resources ResourceRequirements) ResourceRequirements {
	resources.Requests = normalizeResourceList(resources.Requests)
	resources.Limits = normalizeResourceList(resources.Limits)
	return resources
}

func normalizeResourceList(values *ResourceList) *ResourceList {
	if values == nil {
		return nil
	}
	values.CPU = strings.TrimSpace(values.CPU)
	values.Memory = strings.TrimSpace(values.Memory)
	if values.CPU == "" && values.Memory == "" {
		return nil
	}
	return values
}

func mergeResourceRequirements(base, override ResourceRequirements) ResourceRequirements {
	return ResourceRequirements{
		Requests: mergeResourceList(base.Requests, override.Requests),
		Limits:   mergeResourceList(base.Limits, override.Limits),
	}
}

func mergeResourceList(base, override *ResourceList) *ResourceList {
	if base == nil && override == nil {
		return nil
	}
	merged := &ResourceList{}
	if base != nil {
		merged.CPU = strings.TrimSpace(base.CPU)
		merged.Memory = strings.TrimSpace(base.Memory)
	}
	if override != nil {
		if cpu := strings.TrimSpace(override.CPU); cpu != "" {
			merged.CPU = cpu
		}
		if memory := strings.TrimSpace(override.Memory); memory != "" {
			merged.Memory = memory
		}
	}
	if merged.CPU == "" && merged.Memory == "" {
		return nil
	}
	return merged
}

func mergeStringMap(base, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	merged := make(map[string]string, len(base)+len(override))
	for key, value := range base {
		merged[key] = value
	}
	for key, value := range override {
		merged[key] = value
	}
	if len(merged) == 0 {
		return nil
	}
	return merged
}
