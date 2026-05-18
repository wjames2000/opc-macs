package saas

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type FraudDetectionStore struct {
	db *sql.DB
}

func NewFraudDetectionStore(db *sql.DB) *FraudDetectionStore {
	return &FraudDetectionStore{db: db}
}

func (s *FraudDetectionStore) CreateRule(ctx context.Context, rule *FraudRule) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO fraud_rules (id, name, rule_type, config, enabled, severity, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		rule.ID, rule.Name, rule.RuleType, rule.Config, rule.Enabled, rule.Severity,
		rule.CreatedAt, rule.UpdatedAt)
	return err
}

func (s *FraudDetectionStore) ListRules(ctx context.Context) ([]FraudRule, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, rule_type, config, enabled, severity, created_at, updated_at
		 FROM fraud_rules ORDER BY severity DESC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []FraudRule
	for rows.Next() {
		var rule FraudRule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.RuleType, &rule.Config,
			&rule.Enabled, &rule.Severity, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (s *FraudDetectionStore) ToggleRule(ctx context.Context, id string, enabled bool) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE fraud_rules SET enabled = $1, updated_at = $2 WHERE id = $3`,
		enabled, time.Now(), id)
	return err
}

func (s *FraudDetectionStore) LogFlag(ctx context.Context, flag *FraudFlag) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO fraud_flags (id, rule_id, target_type, target_id, reason, severity, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		flag.ID, flag.RuleID, flag.TargetType, flag.TargetID, flag.Reason,
		flag.Severity, flag.Status, flag.CreatedAt)
	return err
}

func (s *FraudDetectionStore) ListFlags(ctx context.Context, status string, limit, offset int) ([]FraudFlag, error) {
	var rows *sql.Rows
	var err error
	if status != "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, rule_id, target_type, target_id, reason, severity, status, created_at
			 FROM fraud_flags WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			status, limit, offset)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, rule_id, target_type, target_id, reason, severity, status, created_at
			 FROM fraud_flags ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []FraudFlag
	for rows.Next() {
		var flag FraudFlag
		if err := rows.Scan(&flag.ID, &flag.RuleID, &flag.TargetType, &flag.TargetID,
			&flag.Reason, &flag.Severity, &flag.Status, &flag.CreatedAt); err != nil {
			return nil, err
		}
		flags = append(flags, flag)
	}
	return flags, rows.Err()
}

type FraudRule struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	RuleType  string    `json:"rule_type"`
	Config    string    `json:"config"`
	Enabled   bool      `json:"enabled"`
	Severity  string    `json:"severity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FraudFlag struct {
	ID         string    `json:"id"`
	RuleID     string    `json:"rule_id"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Reason     string    `json:"reason"`
	Severity   string    `json:"severity"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type FraudDetectionHandler struct {
	store *FraudDetectionStore
}

func NewFraudDetectionHandler(store *FraudDetectionStore) *FraudDetectionHandler {
	return &FraudDetectionHandler{store: store}
}

func (h *FraudDetectionHandler) HandleListRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.store.ListRules(r.Context())
	if err != nil {
		jsonError(w, 500, "list rules: "+err.Error())
		return
	}
	if rules == nil {
		rules = []FraudRule{}
	}
	jsonResponse(w, 200, map[string]interface{}{
		"rules": rules,
		"total": len(rules),
	})
}

func (h *FraudDetectionHandler) HandleCreateRule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		RuleType string `json:"rule_type"`
		Config   string `json:"config"`
		Severity string `json:"severity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if req.Name == "" || req.RuleType == "" {
		jsonError(w, 400, "name and rule_type required")
		return
	}
	if req.Severity == "" {
		req.Severity = "medium"
	}

	rule := &FraudRule{
		ID:        GenerateID(),
		Name:      req.Name,
		RuleType:  req.RuleType,
		Config:    req.Config,
		Enabled:   true,
		Severity:  req.Severity,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.store.CreateRule(r.Context(), rule); err != nil {
		jsonError(w, 500, "create rule: "+err.Error())
		return
	}

	jsonResponse(w, 201, rule)
}

func (h *FraudDetectionHandler) HandleToggleRule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "rule id required")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}

	if err := h.store.ToggleRule(r.Context(), id, req.Enabled); err != nil {
		jsonError(w, 500, "toggle rule: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]interface{}{"id": id, "enabled": req.Enabled})
}

func (h *FraudDetectionHandler) HandleListFlags(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	offset := parseInt(r.URL.Query().Get("offset"), 0)

	flags, err := h.store.ListFlags(r.Context(), status, limit, offset)
	if err != nil {
		jsonError(w, 500, "list flags: "+err.Error())
		return
	}
	if flags == nil {
		flags = []FraudFlag{}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"flags": flags,
		"total": len(flags),
	})
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return defaultVal
		}
		n = n*10 + int(c-'0')
	}
	return n
}
