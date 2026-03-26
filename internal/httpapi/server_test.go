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
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","configMapAnnotations":{"Example.com/config":"true"},"deploymentAnnotations":{"Example.com/deployment":"true"},"podAnnotations":{"Example.com/shared":"true"}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","deploymentAnnotations":{"Example.com/topic-deployment":"true"},"podAnnotations":{"Example.com/topic":"true"}}]}]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	for _, expected := range [][]byte{
		[]byte(`controlPlane.configMapAnnotations key`),
		[]byte(`controlPlane.deploymentAnnotations key`),
		[]byte(`controlPlane.podAnnotations key`),
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
