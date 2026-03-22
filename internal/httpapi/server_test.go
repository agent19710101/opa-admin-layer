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
