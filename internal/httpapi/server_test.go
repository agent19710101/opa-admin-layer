package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agent19710101/opa-admin-layer/internal/admin"
)

func TestValidateEndpoint(t *testing.T) {
	h := NewHandler()
	spec := admin.Specification{
		Name:         "demo",
		ControlPlane: admin.ControlPlane{BaseServiceURL: "https://control.example.com"},
		Tenants:      []admin.Tenant{{Name: "tenant-a", Topics: []admin.Topic{{Name: "billing"}}}},
	}
	body, _ := json.Marshal(spec)
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestValidateEndpointAcceptsMissingContentType(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPlanEndpointRejectsInvalidPayload(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"","tenants":[]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestValidateEndpointRejectsUnknownFields(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","unexpected":true}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("unknown field")) {
		t.Fatalf("expected unknown field error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointAcceptsYAML(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`name: demo
controlPlane:
  baseServiceURL: https://control.example.com
tenants:
  - name: tenant-a
    topics:
      - name: billing
`))
	req.Header.Set("Content-Type", "application/yaml")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"valid":true`)) {
		t.Fatalf("expected valid response, got %s", rec.Body.String())
	}
}

func TestValidateEndpointAcceptsLegacyYAMLContentType(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`name: demo
controlPlane:
  baseServiceURL: https://control.example.com
tenants:
  - name: tenant-a
    topics:
      - name: billing
`))
	req.Header.Set("Content-Type", "application/x-yaml; charset=utf-8")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestValidateEndpointRejectsUnsupportedContentType(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("accepted types are application/json")) {
		t.Fatalf("expected accepted content types in error, got %s", rec.Body.String())
	}
}

func TestPlanEndpointRejectsUnknownYAMLFields(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`name: demo
controlPlane:
  baseServiceURL: https://control.example.com
tenants:
  - name: tenant-a
    topics:
      - name: billing
        unexpected: true
`))
	req.Header.Set("Content-Type", "application/yaml")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("field unexpected not found")) {
		t.Fatalf("expected YAML unknown field error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsInvalidBaseServiceURL(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"/control"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("baseServiceURL")) {
		t.Fatalf("expected invalid baseServiceURL error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsInvalidDefaultListenAddress(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","defaultListenAddress":"localhost"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("defaultListenAddress")) {
		t.Fatalf("expected invalid defaultListenAddress error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsNegativeReplicas(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","replicas":-1},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","replicas":-2}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte("controlPlane.replicas"),
		[]byte(`replicas is invalid`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected invalid replicas error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsInvalidNamespace(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","namespace":"Team-A"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("namespace")) {
		t.Fatalf("expected invalid namespace error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsInvalidTopicLabels(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","labels":{"Example.com/owner":"platform!"}}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("label key")) {
		t.Fatalf("expected invalid label key error, got %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("invalid value")) {
		t.Fatalf("expected invalid label value error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsInvalidRenderedResourceNames(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo!","controlPlane":{"baseServiceURL":"https://control.example.com"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("renders invalid deployment name")) {
		t.Fatalf("expected invalid rendered-name error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsInvalidServiceTypeTrafficPolicyAnnotationKeyAndEmptyOPAResources(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","serviceType":"ExternalName","externalTrafficPolicy":"Edge","internalTrafficPolicy":"Sideways","sessionAffinity":"Sticky","serviceAnnotations":{"Example.com/internal":"true"},"opaResources":{"requests":{}}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("serviceType")) {
		t.Fatalf("expected invalid service type error, got %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("serviceAnnotations key")) {
		t.Fatalf("expected invalid service annotation key error, got %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("externalTrafficPolicy")) {
		t.Fatalf("expected invalid externalTrafficPolicy error, got %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("internalTrafficPolicy")) {
		t.Fatalf("expected invalid internalTrafficPolicy error, got %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("internalTrafficPolicy")) {
		t.Fatalf("expected invalid internalTrafficPolicy error, got %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("sessionAffinity")) {
		t.Fatalf("expected invalid sessionAffinity error, got %s", rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("opaResources.requests")) {
		t.Fatalf("expected invalid opaResources requests error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsInvalidOPAResourceQuantities(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","opaResources":{"requests":{"cpu":"ten millicores","memory":"128Mega"},"limits":{"memory":"0x20"}}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte("opaResources.requests.cpu"),
		[]byte("opaResources.requests.memory"),
		[]byte("opaResources.limits.memory"),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected invalid quantity error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsInvalidTopicOPAResources(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","opaResources":{"requests":{"cpu":"100m","memory":"128Mi"}}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","opaResources":{"requests":{},"limits":{"memory":"0x20"}}}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`opaResources.requests must set cpu and/or memory`),
		[]byte(`opaResources.limits.memory is invalid`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected invalid topic opaResources error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsInvalidTopicServiceOverrides(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","serviceType":"ExternalName","externalTrafficPolicy":"Edge","internalTrafficPolicy":"Sideways","sessionAffinity":"Sticky","serviceAnnotations":{"Example.com/internal":"true"}}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`serviceType is invalid`),
		[]byte(`externalTrafficPolicy is invalid`),
		[]byte(`internalTrafficPolicy is invalid`),
		[]byte(`sessionAffinity is invalid`),
		[]byte(`serviceAnnotations key`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected invalid topic service override error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsOPAResourceRequestsAboveLimits(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","opaResources":{"requests":{"cpu":"1000m"},"limits":{"cpu":"500m"}}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`controlPlane.opaResources.cpu request`),
		[]byte(`effective opaResources.cpu request`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected resource budget validation error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsInvalidInternalTrafficPolicyAndSessionAffinity(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","internalTrafficPolicy":"Sideways","sessionAffinity":"Sticky"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("sessionAffinity")) {
		t.Fatalf("expected invalid sessionAffinity error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsExternalTrafficPolicyWithoutExternallyExposedService(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","externalTrafficPolicy":"Local"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`requires serviceType NodePort or LoadBalancer`)) {
		t.Fatalf("expected externalTrafficPolicy compatibility error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsInvalidConfigMapDeploymentAndPodAnnotationKeys(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","configMapAnnotations":{"Example.com/config":"true"},"deploymentAnnotations":{"Example.com/deployment":"true"},"podAnnotations":{"Example.com/shared":"true"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","configMapAnnotations":{"Example.com/topic-config":"true"},"deploymentAnnotations":{"Example.com/topic-deployment":"true"},"podAnnotations":{"Example.com/topic":"true"}}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`controlPlane.configMapAnnotations key`),
		[]byte(`controlPlane.deploymentAnnotations key`),
		[]byte(`controlPlane.podAnnotations key`),
		[]byte(`configMapAnnotations key \"Example.com/topic-config\" is invalid`),
		[]byte(`deploymentAnnotations key \"Example.com/topic-deployment\" is invalid`),
		[]byte(`podAnnotations key \"Example.com/topic\" is invalid`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected invalid annotation error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestPlanEndpointRendersSharedConfigMapAnnotations(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","configMapAnnotations":{"reloader.stakater.com/match":"true","example.com/source":"generated"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`"configMapManifestYAML":"apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: demo-tenant-a-billing-opa-config\n  annotations:\n    example.com/source: \"generated\"\n    reloader.stakater.com/match: \"true\"`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected rendered config map annotations in response, got %s", rec.Body.String())
		}
	}
}

func TestPlanEndpointRendersMergedTopicConfigMapAnnotations(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","configMapAnnotations":{"reloader.stakater.com/match":"true","example.com/source":"shared"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","configMapAnnotations":{"example.com/source":"billing","example.com/team":"payments"}}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`example.com/source: \"billing\"`),
		[]byte(`example.com/team: \"payments\"`),
		[]byte(`reloader.stakater.com/match: \"true\"`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected rendered merged config map annotation %q in response, got %s", expected, rec.Body.String())
		}
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(`example.com/source: \"shared\"`)) {
		t.Fatalf("expected topic config map annotation override to replace shared value, got %s", rec.Body.String())
	}
}

func TestPlanEndpointRendersMergedTopicConfigMapLabels(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","configMapLabels":{"example.com/config-scope":"shared","example.com/team":"platform"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","configMapLabels":{"example.com/config-scope":"billing","example.com/ring":"canary","app.kubernetes.io/name":"do-not-override"}}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`example.com/config-scope: \"billing\"`),
		[]byte(`example.com/ring: \"canary\"`),
		[]byte(`example.com/team: \"platform\"`),
		[]byte(`app.kubernetes.io/name: \"demo-tenant-a-billing-opa\"`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected rendered config map label %q in response, got %s", expected, rec.Body.String())
		}
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(`app.kubernetes.io/name: \"do-not-override\"`)) {
		t.Fatalf("expected built-in config map label to remain immutable, got %s", rec.Body.String())
	}
}

func TestPlanEndpointRendersMergedTopicServiceLabels(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","serviceLabels":{"example.com/service-scope":"shared","example.com/team":"platform"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","serviceLabels":{"example.com/service-scope":"billing","example.com/ring":"canary","app.kubernetes.io/name":"do-not-override"}}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var plan admin.Plan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("expected valid plan response JSON, got %v body=%s", err, rec.Body.String())
	}
	service := plan.Tenants[0].Topics[0].ServiceManifestYAML
	for _, expected := range []string{
		`example.com/service-scope: "billing"`,
		`example.com/ring: "canary"`,
		`example.com/team: "platform"`,
		`app.kubernetes.io/name: "demo-tenant-a-billing-opa"`,
	} {
		if !bytes.Contains([]byte(service), []byte(expected)) {
			t.Fatalf("expected rendered service label %q in service manifest, got %s", expected, service)
		}
	}
	if bytes.Contains([]byte(service), []byte(`app.kubernetes.io/name: "do-not-override"`)) {
		t.Fatalf("expected built-in service label to remain immutable, got %s", service)
	}
}

func TestPlanEndpointAllowsRemovingInheritedMetadata(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","serviceAnnotations":{"example.com/shared":"true"},"serviceLabels":{"example.com/remove":"true"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","removeServiceAnnotations":["example.com/shared"],"removeServiceLabels":["example.com/remove"]}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var plan admin.Plan
	if err := json.Unmarshal(rec.Body.Bytes(), &plan); err != nil {
		t.Fatalf("expected valid plan response JSON, got %v body=%s", err, rec.Body.String())
	}
	service := plan.Tenants[0].Topics[0].ServiceManifestYAML
	if bytes.Contains([]byte(service), []byte(`example.com/shared`)) || bytes.Contains([]byte(service), []byte(`example.com/remove`)) {
		t.Fatalf("expected inherited service metadata removals to apply, got %s", service)
	}
}

func TestValidateEndpointRejectsInvalidServiceLabels(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","serviceLabels":{"Example.com/shared":"ok","example.com/value":"bad!"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","serviceLabels":{"Example.com/topic":"ok","example.com/ring":"bad!"}}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`controlPlane.serviceLabels key`),
		[]byte(`controlPlane.serviceLabels label`),
		[]byte(`serviceLabels key`),
		[]byte(`serviceLabels label`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected invalid service label error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsInvalidConfigMapLabels(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","configMapLabels":{"Example.com/shared":"ok","example.com/value":"bad!"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","configMapLabels":{"Example.com/topic":"ok","example.com/ring":"bad!"}}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`controlPlane.configMapLabels key`),
		[]byte(`controlPlane.configMapLabels label`),
		[]byte(`configMapLabels key \"Example.com/topic\" is invalid`),
		[]byte(`configMapLabels label \"example.com/ring\" has invalid value`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected invalid config map label error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsInvalidDeploymentLabels(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","deploymentLabels":{"Example.com/shared":"ok","example.com/value":"bad!"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","deploymentLabels":{"Example.com/topic":"ok","example.com/track":"bad!"}}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`controlPlane.deploymentLabels key`),
		[]byte(`controlPlane.deploymentLabels label`),
		[]byte(`deploymentLabels key \"Example.com/topic\" is invalid`),
		[]byte(`deploymentLabels label \"example.com/track\" has invalid value`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected invalid deployment label error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsInvalidServiceAccountName(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","serviceAccountName":"OPA.Shared"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","serviceAccountName":"billing_opa"}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte("controlPlane.serviceAccountName"),
		[]byte(`tenant \"tenant-a\" topic \"billing\" serviceAccountName`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected invalid serviceAccountName error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsInvalidServiceAccountAnnotations(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","serviceAccountAnnotations":{"Example.com/shared":"true"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","serviceAccountAnnotations":{"Example.com/topic":"true"},"removeServiceAccountAnnotations":["bad key"]}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`controlPlane.serviceAccountAnnotations key`),
		[]byte(`serviceAccountAnnotations key \"Example.com/topic\" is invalid`),
		[]byte(`removeServiceAccountAnnotations entry \"bad key\" is invalid`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected invalid serviceAccountAnnotations error %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestPlanEndpointRendersInheritedServiceAccountSettings(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","serviceAccountName":"opa-shared","serviceAccountAnnotations":{"eks.amazonaws.com/role-arn":"arn:aws:iam::123456789012:role/shared-opa"},"automountServiceAccountToken":false},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`serviceAccountName: opa-shared`),
		[]byte(`eks.amazonaws.com/role-arn`),
		[]byte(`arn:aws:iam::123456789012:role/shared-opa`),
		[]byte(`automountServiceAccountToken: false`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected rendered serviceAccount settings to contain %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsAutoscalingWithoutEffectiveCPURequest(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","opaResources":{"requests":{"memory":"128Mi"}},"autoscaling":{"minReplicas":2,"maxReplicas":5,"targetCPUUtilizationPercentage":70}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`effective autoscaling requires effective opaResources.requests.cpu to be set for CPU utilization metrics`)) {
		t.Fatalf("expected autoscaling cpu request error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsConflictingAutoscalingAndReplicas(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","replicas":2,"autoscaling":{"minReplicas":2,"maxReplicas":5,"targetCPUUtilizationPercentage":70}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("controlPlane.replicas is invalid: cannot be set when controlPlane.autoscaling is configured")) {
		t.Fatalf("expected autoscaling conflict error, got %s", rec.Body.String())
	}
}

func TestPlanEndpointRendersAutoscalingManifest(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","opaResources":{"requests":{"cpu":"100m"}},"autoscaling":{"minReplicas":2,"maxReplicas":5,"targetCPUUtilizationPercentage":70}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`"hpaManifestYAML":`),
		[]byte(`kind: HorizontalPodAutoscaler`),
		[]byte(`averageUtilization: 70`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected autoscaling plan response to contain %q, got %s", expected, rec.Body.String())
		}
	}
}

func TestValidateEndpointRejectsAutoscalingWithoutEffectiveMemoryRequest(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","opaResources":{"requests":{"cpu":"100m"}},"autoscaling":{"minReplicas":2,"maxReplicas":5,"targetMemoryUtilizationPercentage":75}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`effective autoscaling requires effective opaResources.requests.memory to be set for memory utilization metrics`)) {
		t.Fatalf("expected autoscaling memory request error, got %s", rec.Body.String())
	}
}

func TestPlanEndpointRendersMemoryAutoscalingMetric(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","opaResources":{"requests":{"memory":"256Mi"}},"autoscaling":{"minReplicas":2,"maxReplicas":5,"targetMemoryUtilizationPercentage":80}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`"hpaManifestYAML":`),
		[]byte(`name: memory`),
		[]byte(`averageUtilization: 80`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected autoscaling plan response to contain %q, got %s", expected, rec.Body.String())
		}
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(`name: cpu`)) {
		t.Fatalf("expected memory-only autoscaling response to omit cpu metric, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsInvalidAutoscalingBehavior(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","opaResources":{"requests":{"cpu":"100m"}},"autoscaling":{"minReplicas":2,"maxReplicas":5,"targetCPUUtilizationPercentage":70,"behavior":{}}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`autoscaling.behavior must configure scaleUp and/or scaleDown`)) {
		t.Fatalf("expected autoscaling behavior validation error, got %s", rec.Body.String())
	}
}

func TestPlanEndpointRendersAutoscalingBehavior(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","opaResources":{"requests":{"cpu":"100m","memory":"256Mi"}},"autoscaling":{"minReplicas":2,"maxReplicas":5,"targetCPUUtilizationPercentage":70,"targetMemoryUtilizationPercentage":80,"behavior":{"scaleUp":{"stabilizationWindowSeconds":30},"scaleDown":{"stabilizationWindowSeconds":300}}}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`behavior:`),
		[]byte(`scaleUp:`),
		[]byte(`stabilizationWindowSeconds: 30`),
		[]byte(`scaleDown:`),
		[]byte(`stabilizationWindowSeconds: 300`),
	} {
		if !bytes.Contains(rec.Body.Bytes(), expected) {
			t.Fatalf("expected autoscaling behavior in plan response %q, got %s", expected, rec.Body.String())
		}
	}
}
