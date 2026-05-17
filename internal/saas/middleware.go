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

func AuthMiddleware(store *UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") || len(auth) < 30 {
				// Try X-User-ID header fallback
				userID := r.Header.Get("X-User-ID")
				role := r.Header.Get("X-User-Role")
				if userID != "" {
					ctx := context.WithValue(r.Context(), CtxUserID, userID)
					if role != "" {
						ctx = context.WithValue(ctx, CtxUserRole, role)
					}
					r = r.WithContext(ctx)
				}
				next.ServeHTTP(w, r)
				return
			}

			token := strings.TrimPrefix(auth, "Bearer ")
			if strings.HasPrefix(token, "tk_") {
				if user, err := store.GetByID(r.Context(), token); err == nil {
					ctx := context.WithValue(r.Context(), CtxUserID, user.ID)
					if user.Role != "" {
						ctx = context.WithValue(ctx, CtxUserRole, user.Role)
					}
					r = r.WithContext(ctx)
				}
			} else {
				// Direct user ID as token
				if user, err := store.GetByID(r.Context(), token); err == nil {
					ctx := context.WithValue(r.Context(), CtxUserID, user.ID)
					if user.Role != "" {
						ctx = context.WithValue(ctx, CtxUserRole, user.Role)
					}
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(CtxUserRole).(string)
			if role == "" {
				jsonError(w, 403, "forbidden: no role assigned")
				return
			}
			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}
			jsonError(w, 403, "forbidden: insufficient role")
		})
	}
}

func getUserRole(r *http.Request) string {
	if role, ok := r.Context().Value(CtxUserRole).(string); ok {
		return role
	}
	return ""
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	jsonResponse(w, status, map[string]string{"error": msg})
}
