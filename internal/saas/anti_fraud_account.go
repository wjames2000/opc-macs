package saas

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type FraudAccountStore struct {
	db *sql.DB
}

func NewFraudAccountStore(db *sql.DB) *FraudAccountStore {
	return &FraudAccountStore{db: db}
}

func (s *FraudAccountStore) CheckDuplicateAccount(ctx context.Context, tenantID, ipAddress, deviceID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM accounts
		 WHERE tenant_id = $1 AND (ip_address = $2 OR device_id = $3)
		 AND created_at > $4`,
		tenantID, ipAddress, deviceID, time.Now().Add(-24*time.Hour)).Scan(&count)
	return count, err
}

func (s *FraudAccountStore) LogSuspiciousActivity(ctx context.Context, act *SuspiciousActivity) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO suspicious_activities (id, tenant_id, user_id, activity_type, ip_address, device_id, reason, severity, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		act.ID, act.TenantID, act.UserID, act.ActivityType, act.IPAddress,
		act.DeviceID, act.Reason, act.Severity, act.CreatedAt)
	return err
}

func (s *FraudAccountStore) ListSuspiciousActivities(ctx context.Context, tenantID string, limit, offset int) ([]SuspiciousActivity, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, user_id, activity_type, ip_address, device_id, reason, severity, created_at
		 FROM suspicious_activities WHERE tenant_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var acts []SuspiciousActivity
	for rows.Next() {
		var act SuspiciousActivity
		if err := rows.Scan(&act.ID, &act.TenantID, &act.UserID, &act.ActivityType,
			&act.IPAddress, &act.DeviceID, &act.Reason, &act.Severity, &act.CreatedAt); err != nil {
			return nil, err
		}
		acts = append(acts, act)
	}
	return acts, rows.Err()
}

type SuspiciousActivity struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	UserID       string    `json:"user_id"`
	ActivityType string    `json:"activity_type"`
	IPAddress    string    `json:"ip_address"`
	DeviceID     string    `json:"device_id"`
	Reason       string    `json:"reason"`
	Severity     string    `json:"severity"`
	CreatedAt    time.Time `json:"created_at"`
}

type FraudAccountHandler struct {
	store *FraudAccountStore
}

func NewFraudAccountHandler(store *FraudAccountStore) *FraudAccountHandler {
	return &FraudAccountHandler{store: store}
}

func (h *FraudAccountHandler) HandleCheckDuplicate(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := getTenantID(r), getUserID(r)

	ipAddress := r.URL.Query().Get("ip_address")
	deviceID := r.URL.Query().Get("device_id")
	if ipAddress == "" && deviceID == "" {
		jsonError(w, 400, "ip_address or device_id required")
		return
	}

	count, err := h.store.CheckDuplicateAccount(r.Context(), tenantID, ipAddress, deviceID)
	if err != nil {
		jsonError(w, 500, "check duplicate: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]interface{}{
		"duplicate_count": count,
		"suspicious":      count >= 3,
	})
}

func (h *FraudAccountHandler) HandleReportActivity(w http.ResponseWriter, r *http.Request) {
	tenantID, userID := getTenantID(r), getUserID(r)

	var req struct {
		ActivityType string `json:"activity_type"`
		IPAddress    string `json:"ip_address"`
		DeviceID     string `json:"device_id"`
		Reason       string `json:"reason"`
		Severity     string `json:"severity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if req.ActivityType == "" {
		jsonError(w, 400, "activity_type required")
		return
	}
	if req.Severity == "" {
		req.Severity = "low"
	}

	act := &SuspiciousActivity{
		ID:           GenerateID(),
		TenantID:     tenantID,
		UserID:       userID,
		ActivityType: req.ActivityType,
		IPAddress:    req.IPAddress,
		DeviceID:     req.DeviceID,
		Reason:       req.Reason,
		Severity:     req.Severity,
		CreatedAt:    time.Now(),
	}

	if err := h.store.LogSuspiciousActivity(r.Context(), act); err != nil {
		jsonError(w, 500, "log activity: "+err.Error())
		return
	}

	jsonResponse(w, 201, act)
}

func (h *FraudAccountHandler) HandleListActivities(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := getTenantID(r), getUserID(r)
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	offset := parseInt(r.URL.Query().Get("offset"), 0)

	acts, err := h.store.ListSuspiciousActivities(r.Context(), tenantID, limit, offset)
	if err != nil {
		jsonError(w, 500, "list activities: "+err.Error())
		return
	}
	if acts == nil {
		acts = []SuspiciousActivity{}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"activities": acts,
		"total":      len(acts),
	})
}
