package saas

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type contextKey string

const (
	CtxTenantID contextKey = "tenant_id"
	CtxUserID   contextKey = "user_id"
	CtxUserRole contextKey = "user_role"
)

func TenantMiddleware(store *TenantStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tenantID string

			// 1. Try API Key header
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer tk_") {
				apiKey := strings.TrimPrefix(auth, "Bearer ")
				user, err := NewUserStore(store.db).GetByID(r.Context(), apiKey)
				if err == nil {
					tenantID = user.TenantID
				}
			}

			// 2. Try X-Tenant-ID header
			if tenantID == "" {
				tenantID = r.Header.Get("X-Tenant-ID")
			}

			// 3. Try subdomain
			if tenantID == "" {
				host := r.Host
				if parts := strings.SplitN(host, ".", 2); len(parts) == 2 {
					slug := parts[0]
					if t, err := store.GetByID(r.Context(), slug); err == nil {
						tenantID = t.ID
					}
				}
			}

			if tenantID != "" {
				ctx := context.WithValue(r.Context(), CtxTenantID, tenantID)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	jsonResponse(w, status, map[string]string{"error": msg})
}
