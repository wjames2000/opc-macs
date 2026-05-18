package saas

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func getTenantID(r *http.Request) string {
	if v := r.Context().Value(CtxTenantID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return r.Header.Get("X-Tenant-ID")
}

func getUserID(r *http.Request) string {
	if v := r.Context().Value(CtxUserID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return r.Header.Get("X-User-ID")
}
