package saas

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type FraudAlertStore struct {
	db *sql.DB
}

func NewFraudAlertStore(db *sql.DB) *FraudAlertStore {
	return &FraudAlertStore{db: db}
}

func (s *FraudAlertStore) CreateAlert(ctx context.Context, alert *FraudAlert) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO fraud_alerts (id, tenant_id, rule_id, flag_id, title, description,
		 severity, status, assigned_to, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		alert.ID, alert.TenantID, alert.RuleID, alert.FlagID, alert.Title,
		alert.Description, alert.Severity, alert.Status, alert.AssignedTo,
		alert.CreatedAt, alert.UpdatedAt)
	return err
}

func (s *FraudAlertStore) ListAlerts(ctx context.Context, tenantID, status string, limit, offset int) ([]FraudAlert, error) {
	var rows *sql.Rows
	var err error
	if status != "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, tenant_id, rule_id, flag_id, title, description,
			 severity, status, assigned_to, created_at, updated_at
			 FROM fraud_alerts WHERE tenant_id = $1 AND status = $2
			 ORDER BY created_at DESC LIMIT $3 OFFSET $4`,
			tenantID, status, limit, offset)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, tenant_id, rule_id, flag_id, title, description,
			 severity, status, assigned_to, created_at, updated_at
			 FROM fraud_alerts WHERE tenant_id = $1
			 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			tenantID, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []FraudAlert
	for rows.Next() {
		var alert FraudAlert
		if err := rows.Scan(&alert.ID, &alert.TenantID, &alert.RuleID, &alert.FlagID,
			&alert.Title, &alert.Description, &alert.Severity, &alert.Status,
			&alert.AssignedTo, &alert.CreatedAt, &alert.UpdatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, rows.Err()
}

func (s *FraudAlertStore) UpdateAlertStatus(ctx context.Context, id, status, assignedTo string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE fraud_alerts SET status = $1, assigned_to = $2, updated_at = $3 WHERE id = $4`,
		status, assignedTo, time.Now(), id)
	return err
}

func (s *FraudAlertStore) GetAlertCountByStatus(ctx context.Context, tenantID string) (map[string]int, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT status, COUNT(*) FROM fraud_alerts WHERE tenant_id = $1 GROUP BY status`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}
	return counts, rows.Err()
}

type FraudAlert struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	RuleID      string    `json:"rule_id,omitempty"`
	FlagID      string    `json:"flag_id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	AssignedTo  string    `json:"assigned_to,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FraudAlertHandler struct {
	store *FraudAlertStore
}

func NewFraudAlertHandler(store *FraudAlertStore) *FraudAlertHandler {
	return &FraudAlertHandler{store: store}
}

func (h *FraudAlertHandler) HandleListAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := getTenantID(r), getUserID(r)
	status := r.URL.Query().Get("status")
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	offset := parseInt(r.URL.Query().Get("offset"), 0)

	alerts, err := h.store.ListAlerts(r.Context(), tenantID, status, limit, offset)
	if err != nil {
		jsonError(w, 500, "list alerts: "+err.Error())
		return
	}
	if alerts == nil {
		alerts = []FraudAlert{}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

func (h *FraudAlertHandler) HandleCreateAlert(w http.ResponseWriter, r *http.Request) {
	tenantID, userID := getTenantID(r), getUserID(r)

	var req struct {
		RuleID      string `json:"rule_id"`
		FlagID      string `json:"flag_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Severity    string `json:"severity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if req.Title == "" {
		jsonError(w, 400, "title required")
		return
	}
	if req.Severity == "" {
		req.Severity = "medium"
	}

	alert := &FraudAlert{
		ID:          GenerateID(),
		TenantID:    tenantID,
		RuleID:      req.RuleID,
		FlagID:      req.FlagID,
		Title:       req.Title,
		Description: req.Description,
		Severity:    req.Severity,
		Status:      "open",
		AssignedTo:  userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.store.CreateAlert(r.Context(), alert); err != nil {
		jsonError(w, 500, "create alert: "+err.Error())
		return
	}

	jsonResponse(w, 201, alert)
}

func (h *FraudAlertHandler) HandleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "alert id required")
		return
	}

	var req struct {
		Status     string `json:"status"`
		AssignedTo string `json:"assigned_to,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if req.Status == "" {
		jsonError(w, 400, "status required")
		return
	}

	if err := h.store.UpdateAlertStatus(r.Context(), id, req.Status, req.AssignedTo); err != nil {
		jsonError(w, 500, "update alert: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]string{"status": req.Status})
}

func (h *FraudAlertHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := getTenantID(r), getUserID(r)

	counts, err := h.store.GetAlertCountByStatus(r.Context(), tenantID)
	if err != nil {
		jsonError(w, 500, "get alert stats: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]interface{}{
		"stats": counts,
	})
}
