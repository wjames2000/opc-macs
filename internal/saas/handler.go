package saas

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type Server struct {
	Tenants    *TenantStore
	Users      *UserStore
	Loader     *runtime.Loader
	ModelKey   string // fallback model API key
	ModelURL   string // fallback model base URL
}

func NewServer(db *sql.DB, loader *runtime.Loader) (*Server, error) {
	tenants := NewTenantStore(db)
	users := NewUserStore(db)

	if err := tenants.InitSchema(); err != nil {
		return nil, err
	}

	return &Server{
		Tenants: tenants,
		Users:   users,
		Loader:  loader,
	}, nil
}

// POST /api/v1/auth/register
func (s *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}
	var req struct {
		TenantName string `json:"tenant_name"`
		Slug       string `json:"slug"`
		Email      string `json:"email"`
		Password   string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		jsonError(w, 400, "email and password required")
		return
	}

	// Create tenant and owner user
	tenant, err := s.Tenants.Create(r.Context(), req.TenantName, req.Slug)
	if err != nil {
		jsonError(w, 409, "tenant creation failed: "+err.Error())
		return
	}
	user, err := s.Users.Create(r.Context(), tenant.ID, req.Email, req.Password)
	if err != nil {
		jsonError(w, 409, "user creation failed: "+err.Error())
		return
	}

	jsonResponse(w, 201, map[string]interface{}{
		"tenant_id": tenant.ID,
		"user_id":   user.ID,
		"api_key":   user.ID, // simple: use user ID as API key
	})
}

// POST /api/v1/auth/login
func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request")
		return
	}
	user, err := s.Users.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		jsonError(w, 401, "invalid credentials")
		return
	}
	token := GenerateToken(user.ID)
	jsonResponse(w, 200, map[string]interface{}{
		"token":     token,
		"user_id":   user.ID,
		"tenant_id": user.TenantID,
		"role":      user.Role,
	})
}

// GET /api/v1/agents
func (s *Server) HandleListAgents(w http.ResponseWriter, r *http.Request) {
	plugins := s.Loader.List()
	jsonResponse(w, 200, plugins)
}

// GET /api/v1/tenants
func (s *Server) HandleListTenants(w http.ResponseWriter, r *http.Request) {
	tenants, err := s.Tenants.GetAll(r.Context())
	if err != nil {
		jsonError(w, 500, "failed to list tenants")
		return
	}
	jsonResponse(w, 200, tenants)
}

// GET /api/v1/healthz
func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, map[string]string{"status": "ok", "service": "opc-agent-saas"})
}
