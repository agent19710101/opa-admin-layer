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
	BaseServiceURL       string               `json:"baseServiceURL" yaml:"baseServiceURL"`
	BundlePrefix         string               `json:"bundlePrefix" yaml:"bundlePrefix"`
	DefaultDecisionPath  string               `json:"defaultDecisionPath" yaml:"defaultDecisionPath"`
	DefaultListenAddress string               `json:"defaultListenAddress" yaml:"defaultListenAddress"`
	OPAImage             string               `json:"opaImage" yaml:"opaImage"`
	ServiceType          string               `json:"serviceType" yaml:"serviceType"`
	ServiceAnnotations   map[string]string    `json:"serviceAnnotations" yaml:"serviceAnnotations"`
	OPAResources         ResourceRequirements `json:"opaResources" yaml:"opaResources"`
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
	Name           string            `json:"name" yaml:"name"`
	BundleResource string            `json:"bundleResource,omitempty" yaml:"bundleResource,omitempty"`
	DecisionPath   string            `json:"decisionPath,omitempty" yaml:"decisionPath,omitempty"`
	Labels         map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
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
	seenTenants := map[string]struct{}{}
	if err := validateKubernetesServiceType(spec.ControlPlane.ServiceType); err != nil {
		issues = append(issues, fmt.Sprintf("controlPlane.serviceType is invalid: %v", err))
	}
	for annotationKey := range spec.ControlPlane.ServiceAnnotations {
		if err := validateKubernetesLabelKey(annotationKey); err != nil {
			issues = append(issues, fmt.Sprintf("controlPlane.serviceAnnotations key %q is invalid: %v", annotationKey, err))
		}
	}
	issues = append(issues, validateOPAResources(spec.ControlPlane.OPAResources)...)
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
			for labelKey, labelValue := range topic.Labels {
				if err := validateKubernetesLabelKey(labelKey); err != nil {
					issues = append(issues, fmt.Sprintf("tenant %q topic %q label key %q is invalid: %v", tenantName, topicName, labelKey, err))
				}
				if err := validateKubernetesLabelValue(labelValue); err != nil {
					issues = append(issues, fmt.Sprintf("tenant %q topic %q label %q has invalid value %q: %v", tenantName, topicName, labelKey, labelValue, err))
				}
			}
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
	var issues []string
	issues = append(issues, validateResourceList("controlPlane.opaResources.requests", resources.Requests)...)
	issues = append(issues, validateResourceList("controlPlane.opaResources.limits", resources.Limits)...)
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
	if strings.TrimSpace(normalized.ControlPlane.ServiceType) == "" {
		normalized.ControlPlane.ServiceType = "ClusterIP"
	}
	normalized.ControlPlane.OPAResources = normalizeResourceRequirements(normalized.ControlPlane.OPAResources)
	for i := range normalized.Tenants {
		normalized.Tenants[i].Name = strings.TrimSpace(normalized.Tenants[i].Name)
		for j := range normalized.Tenants[i].Topics {
			normalized.Tenants[i].Topics[j].Name = strings.TrimSpace(normalized.Tenants[i].Topics[j].Name)
			if strings.TrimSpace(normalized.Tenants[i].Topics[j].DecisionPath) == "" {
				normalized.Tenants[i].Topics[j].DecisionPath = normalized.ControlPlane.DefaultDecisionPath
			}
			if strings.TrimSpace(normalized.Tenants[i].Topics[j].BundleResource) == "" {
				normalized.Tenants[i].Topics[j].BundleResource = fmt.Sprintf("%s/%s/%s.tar.gz", strings.Trim(normalized.ControlPlane.BundlePrefix, "/"), normalized.Tenants[i].Name, normalized.Tenants[i].Topics[j].Name)
			}
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
