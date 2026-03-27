package admin

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeSpecAcceptsYAML(t *testing.T) {
	payload := []byte(`name: demo
controlPlane:
  baseServiceURL: https://control.example.com
  namespace: policy-system
  replicas: 3
  serviceAnnotations:
    example.com/internal: "true"
  serviceLabels:
    example.com/service-scope: shared
  configMapAnnotations:
    reloader.stakater.com/match: "true"
  configMapLabels:
    example.com/config-scope: shared
  deploymentAnnotations:
    example.com/owner: platform
  podAnnotations:
    sidecar.istio.io/inject: "false"
  podLabels:
    example.com/workload-class: shared
  serviceAccountName: opa-shared
  serviceAccountAnnotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/shared-opa
  serviceAccountLabels:
    example.com/service-account-scope: shared
  automountServiceAccountToken: false
tenants:
  - name: tenant-a
    topics:
      - name: billing
        replicas: 5
        serviceAccountName: billing-opa
        serviceAccountAnnotations:
          eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/billing-opa
        removeServiceAccountAnnotations:
          - eks.amazonaws.com/role-arn
        serviceAccountLabels:
          example.com/service-account-scope: topic
        removeServiceAccountLabels:
          - example.com/service-account-scope
        automountServiceAccountToken: true
        serviceLabels:
          example.com/service-scope: topic
        removeServiceAnnotations:
          - example.com/internal
        configMapAnnotations:
          example.com/source: topic
        removeConfigMapLabels:
          - example.com/config-scope
        configMapLabels:
          example.com/config-scope: topic
        podLabels:
          example.com/workload-class: topic
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
	if spec.ControlPlane.Namespace != "policy-system" {
		t.Fatalf("unexpected namespace: %q", spec.ControlPlane.Namespace)
	}
	if spec.ControlPlane.Replicas != 3 {
		t.Fatalf("unexpected replicas: %d", spec.ControlPlane.Replicas)
	}
	if got := spec.ControlPlane.ServiceAnnotations["example.com/internal"]; got != "true" {
		t.Fatalf("expected service annotation to decode, got %q", got)
	}
	if got := spec.ControlPlane.ServiceLabels["example.com/service-scope"]; got != "shared" {
		t.Fatalf("expected service label to decode, got %q", got)
	}
	if got := spec.ControlPlane.ConfigMapAnnotations["reloader.stakater.com/match"]; got != "true" {
		t.Fatalf("expected config map annotation to decode, got %q", got)
	}
	if got := spec.ControlPlane.ConfigMapLabels["example.com/config-scope"]; got != "shared" {
		t.Fatalf("expected config map label to decode, got %q", got)
	}
	if got := spec.ControlPlane.DeploymentAnnotations["example.com/owner"]; got != "platform" {
		t.Fatalf("expected deployment annotation to decode, got %q", got)
	}
	if got := spec.ControlPlane.PodAnnotations["sidecar.istio.io/inject"]; got != "false" {
		t.Fatalf("expected pod annotation to decode, got %q", got)
	}
	if got := spec.ControlPlane.PodLabels["example.com/workload-class"]; got != "shared" {
		t.Fatalf("expected pod label to decode, got %q", got)
	}
	if spec.ControlPlane.ServiceAccountName != "opa-shared" {
		t.Fatalf("expected shared serviceAccountName to decode, got %q", spec.ControlPlane.ServiceAccountName)
	}
	if got := spec.ControlPlane.ServiceAccountAnnotations["eks.amazonaws.com/role-arn"]; got != "arn:aws:iam::123456789012:role/shared-opa" {
		t.Fatalf("expected shared serviceAccountAnnotations to decode, got %q", got)
	}
	if got := spec.ControlPlane.ServiceAccountLabels["example.com/service-account-scope"]; got != "shared" {
		t.Fatalf("expected shared serviceAccountLabels to decode, got %q", got)
	}
	if spec.ControlPlane.AutomountServiceAccountToken == nil || *spec.ControlPlane.AutomountServiceAccountToken {
		t.Fatalf("expected shared automountServiceAccountToken=false to decode, got %#v", spec.ControlPlane.AutomountServiceAccountToken)
	}
	if len(spec.Tenants) != 1 || len(spec.Tenants[0].Topics) != 1 || spec.Tenants[0].Topics[0].Name != "billing" {
		t.Fatalf("unexpected tenant/topic decode result: %#v", spec.Tenants)
	}
	if spec.Tenants[0].Topics[0].Replicas != 5 {
		t.Fatalf("expected topic replicas to decode, got %d", spec.Tenants[0].Topics[0].Replicas)
	}
	if spec.Tenants[0].Topics[0].ServiceAccountName != "billing-opa" {
		t.Fatalf("expected topic serviceAccountName to decode, got %q", spec.Tenants[0].Topics[0].ServiceAccountName)
	}
	if got := spec.Tenants[0].Topics[0].ServiceAccountAnnotations["eks.amazonaws.com/role-arn"]; got != "arn:aws:iam::123456789012:role/billing-opa" {
		t.Fatalf("expected topic serviceAccountAnnotations to decode, got %q", got)
	}
	if got := spec.Tenants[0].Topics[0].RemoveServiceAccountAnnotations; !reflect.DeepEqual(got, []string{"eks.amazonaws.com/role-arn"}) {
		t.Fatalf("expected topic removeServiceAccountAnnotations to decode, got %#v", got)
	}
	if got := spec.Tenants[0].Topics[0].ServiceAccountLabels["example.com/service-account-scope"]; got != "topic" {
		t.Fatalf("expected topic serviceAccountLabels to decode, got %q", got)
	}
	if got := spec.Tenants[0].Topics[0].RemoveServiceAccountLabels; !reflect.DeepEqual(got, []string{"example.com/service-account-scope"}) {
		t.Fatalf("expected topic removeServiceAccountLabels to decode, got %#v", got)
	}
	if spec.Tenants[0].Topics[0].AutomountServiceAccountToken == nil || !*spec.Tenants[0].Topics[0].AutomountServiceAccountToken {
		t.Fatalf("expected topic automountServiceAccountToken=true to decode, got %#v", spec.Tenants[0].Topics[0].AutomountServiceAccountToken)
	}
	if got := spec.Tenants[0].Topics[0].ServiceLabels["example.com/service-scope"]; got != "topic" {
		t.Fatalf("expected topic serviceLabels to decode, got %q", got)
	}
	if got := spec.Tenants[0].Topics[0].RemoveServiceAnnotations; !reflect.DeepEqual(got, []string{"example.com/internal"}) {
		t.Fatalf("expected topic removeServiceAnnotations to decode, got %#v", got)
	}
	if got := spec.Tenants[0].Topics[0].ConfigMapAnnotations["example.com/source"]; got != "topic" {
		t.Fatalf("expected topic configMapAnnotations to decode, got %q", got)
	}
	if got := spec.Tenants[0].Topics[0].RemoveConfigMapLabels; !reflect.DeepEqual(got, []string{"example.com/config-scope"}) {
		t.Fatalf("expected topic removeConfigMapLabels to decode, got %#v", got)
	}
	if got := spec.Tenants[0].Topics[0].ConfigMapLabels["example.com/config-scope"]; got != "topic" {
		t.Fatalf("expected topic configMapLabels to decode, got %q", got)
	}
	if got := spec.Tenants[0].Topics[0].PodLabels["example.com/workload-class"]; got != "topic" {
		t.Fatalf("expected topic pod label to decode, got %q", got)
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

func TestLoadSpecExamplesRemainEquivalentAcrossJSONAndYAML(t *testing.T) {
	jsonSpec, err := LoadSpec(filepath.Join("..", "..", "deploy", "examples", "dev-spec.json"))
	if err != nil {
		t.Fatalf("load JSON example: %v", err)
	}
	yamlSpec, err := LoadSpec(filepath.Join("..", "..", "deploy", "examples", "dev-spec.yaml"))
	if err != nil {
		t.Fatalf("load YAML example: %v", err)
	}

	if !reflect.DeepEqual(normalize(jsonSpec), normalize(yamlSpec)) {
		t.Fatalf("expected checked-in JSON and YAML examples to stay equivalent\njson: %#v\nyaml: %#v", normalize(jsonSpec), normalize(yamlSpec))
	}
}

func TestDecodeSpecAcceptsAutoscalingYAML(t *testing.T) {
	payload := []byte(`name: demo
controlPlane:
  baseServiceURL: https://control.example.com
  autoscaling:
    minReplicas: 2
    maxReplicas: 5
    targetCPUUtilizationPercentage: 70
    behavior:
      scaleUp:
        selectPolicy: Max
        policies:
          - type: Pods
            value: 2
            periodSeconds: 60
tenants:
  - name: tenant-a
    topics:
      - name: billing
        autoscaling:
          minReplicas: 3
          maxReplicas: 7
          targetCPUUtilizationPercentage: 80
          behavior:
            scaleDown:
              selectPolicy: Min
              policies:
                - type: Percent
                  value: 25
                  periodSeconds: 120
`)

	spec, err := DecodeSpec(payload)
	if err != nil {
		t.Fatalf("DecodeSpec returned error: %v", err)
	}
	if spec.ControlPlane.Autoscaling == nil || spec.ControlPlane.Autoscaling.MinReplicas != 2 || spec.ControlPlane.Autoscaling.MaxReplicas != 5 || spec.ControlPlane.Autoscaling.TargetCPUUtilizationPercentage != 70 {
		t.Fatalf("expected shared autoscaling to decode, got %#v", spec.ControlPlane.Autoscaling)
	}
	if spec.ControlPlane.Autoscaling.Behavior == nil || spec.ControlPlane.Autoscaling.Behavior.ScaleUp == nil || spec.ControlPlane.Autoscaling.Behavior.ScaleUp.SelectPolicy != "Max" {
		t.Fatalf("expected shared autoscaling behavior to decode, got %#v", spec.ControlPlane.Autoscaling)
	}
	if got := spec.ControlPlane.Autoscaling.Behavior.ScaleUp.Policies; !reflect.DeepEqual(got, []HPAScalingPolicy{{Type: "Pods", Value: 2, PeriodSeconds: 60}}) {
		t.Fatalf("expected shared autoscaling policies to decode, got %#v", got)
	}
	if spec.Tenants[0].Topics[0].Autoscaling == nil || spec.Tenants[0].Topics[0].Autoscaling.MinReplicas != 3 || spec.Tenants[0].Topics[0].Autoscaling.MaxReplicas != 7 || spec.Tenants[0].Topics[0].Autoscaling.TargetCPUUtilizationPercentage != 80 {
		t.Fatalf("expected topic autoscaling to decode, got %#v", spec.Tenants[0].Topics[0].Autoscaling)
	}
	if spec.Tenants[0].Topics[0].Autoscaling.Behavior == nil || spec.Tenants[0].Topics[0].Autoscaling.Behavior.ScaleDown == nil || spec.Tenants[0].Topics[0].Autoscaling.Behavior.ScaleDown.SelectPolicy != "Min" {
		t.Fatalf("expected topic autoscaling behavior to decode, got %#v", spec.Tenants[0].Topics[0].Autoscaling)
	}
	if got := spec.Tenants[0].Topics[0].Autoscaling.Behavior.ScaleDown.Policies; !reflect.DeepEqual(got, []HPAScalingPolicy{{Type: "Percent", Value: 25, PeriodSeconds: 120}}) {
		t.Fatalf("expected topic autoscaling policies to decode, got %#v", got)
	}
}
