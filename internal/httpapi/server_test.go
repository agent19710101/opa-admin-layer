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
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPlanEndpointRejectsInvalidPayload(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/plans", bytes.NewBufferString(`{"name":"","tenants":[]}`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestValidateEndpointRejectsUnknownFields(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","unexpected":true}]}]}`))
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
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte("defaultListenAddress")) {
		t.Fatalf("expected invalid defaultListenAddress error, got %s", rec.Body.String())
	}
}

func TestValidateEndpointRejectsInvalidTopicLabels(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com"},"tenants":[{"name":"tenant-a","topics":[{"name":"billing","labels":{"Example.com/owner":"platform!"}}]}]}`))
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

func TestValidateEndpointRejectsInvalidServiceTypeAnnotationKeyAndEmptyOPAResources(t *testing.T) {
	h := NewHandler()
	req := httptest.NewRequest(http.MethodPost, "/v1/validate", bytes.NewBufferString(`{"name":"demo","controlPlane":{"baseServiceURL":"https://control.example.com","serviceType":"ExternalName","serviceAnnotations":{"Example.com/internal":"true"},"opaResources":{"requests":{}}},"tenants":[{"name":"tenant-a","topics":[{"name":"billing"}]}]}`))
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
