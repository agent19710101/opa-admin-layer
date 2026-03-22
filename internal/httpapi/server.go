package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/agent19710101/opa-admin-layer/internal/admin"
)

func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, NewHandler())
}

func NewHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})
	mux.HandleFunc("/v1/validate", func(w http.ResponseWriter, r *http.Request) {
		spec, ok := decodeSpec(w, r)
		if !ok {
			return
		}
		issues := admin.Validate(spec)
		status := http.StatusOK
		if len(issues) > 0 {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]any{
			"valid":  len(issues) == 0,
			"issues": issues,
		})
	})
	mux.HandleFunc("/v1/plans", func(w http.ResponseWriter, r *http.Request) {
		spec, ok := decodeSpec(w, r)
		if !ok {
			return
		}
		plan, err := admin.BuildPlan(spec)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, plan)
	})
	return mux
}

func decodeSpec(w http.ResponseWriter, r *http.Request) (admin.Specification, bool) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": fmt.Sprintf("method %s not allowed", r.Method)})
		return admin.Specification{}, false
	}
	defer r.Body.Close()
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return admin.Specification{}, false
	}
	spec, err := admin.DecodeSpec(payload)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return admin.Specification{}, false
	}
	return spec, true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
