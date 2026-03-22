package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Specification struct {
	Name         string       `json:"name"`
	ControlPlane ControlPlane `json:"controlPlane"`
	Tenants      []Tenant     `json:"tenants"`
}

type ControlPlane struct {
	BaseServiceURL       string `json:"baseServiceURL"`
	BundlePrefix         string `json:"bundlePrefix"`
	DefaultDecisionPath  string `json:"defaultDecisionPath"`
	DefaultListenAddress string `json:"defaultListenAddress"`
}

type Tenant struct {
	Name   string  `json:"name"`
	Topics []Topic `json:"topics"`
}

type Topic struct {
	Name           string            `json:"name"`
	BundleResource string            `json:"bundleResource,omitempty"`
	DecisionPath   string            `json:"decisionPath,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
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

// DecodeSpec decodes a specification from JSON and rejects unknown fields.
func DecodeSpec(payload []byte) (Specification, error) {
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()

	var spec Specification
	if err := dec.Decode(&spec); err != nil {
		return Specification{}, err
	}
	if err := ensureSingleJSONValue(dec); err != nil {
		return Specification{}, err
	}
	return spec, nil
}

func ensureSingleJSONValue(dec *json.Decoder) error {
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("spec must contain exactly one JSON object")
		}
		return err
	}
	return nil
}

func Validate(spec Specification) []string {
	var issues []string
	if strings.TrimSpace(spec.Name) == "" {
		issues = append(issues, "spec.name must not be empty")
	}
	if strings.TrimSpace(spec.ControlPlane.BaseServiceURL) == "" {
		issues = append(issues, "controlPlane.baseServiceURL must not be empty")
	}
	seenTenants := map[string]struct{}{}
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
		}
	}
	sort.Strings(issues)
	return issues
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
