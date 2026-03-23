package admin

import (
	"strings"
	"testing"
)

func TestDecodeSpecAcceptsYAML(t *testing.T) {
	payload := []byte(`name: demo
controlPlane:
  baseServiceURL: https://control.example.com
  serviceAnnotations:
    example.com/internal: "true"
tenants:
  - name: tenant-a
    topics:
      - name: billing
`)

	spec, err := DecodeSpec(payload)
	if err != nil {
		t.Fatalf("DecodeSpec returned error: %v", err)
	}
	if spec.Name != "demo" {
		t.Fatalf("expected spec name demo, got %q", spec.Name)
	}
	if spec.ControlPlane.BaseServiceURL != "https://control.example.com" {
		t.Fatalf("unexpected baseServiceURL: %q", spec.ControlPlane.BaseServiceURL)
	}
	if got := spec.ControlPlane.ServiceAnnotations["example.com/internal"]; got != "true" {
		t.Fatalf("expected service annotation to decode, got %q", got)
	}
	if len(spec.Tenants) != 1 || len(spec.Tenants[0].Topics) != 1 || spec.Tenants[0].Topics[0].Name != "billing" {
		t.Fatalf("unexpected tenant/topic decode result: %#v", spec.Tenants)
	}
}

func TestDecodeSpecRejectsUnknownYAMLFields(t *testing.T) {
	payload := []byte(`name: demo
controlPlane:
  baseServiceURL: https://control.example.com
tenants:
  - name: tenant-a
    topics:
      - name: billing
        unexpected: true
`)

	_, err := DecodeSpec(payload)
	if err == nil {
		t.Fatal("expected DecodeSpec to reject unknown YAML field")
	}
	if !strings.Contains(err.Error(), "field unexpected not found") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}
