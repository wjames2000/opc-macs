package saas

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type FraudContentStore struct {
	db *sql.DB
}

func NewFraudContentStore(db *sql.DB) *FraudContentStore {
	return &FraudContentStore{db: db}
}

func (s *FraudContentStore) CheckDuplicateContent(ctx context.Context, contentHash string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM content_checks WHERE content_hash = $1 AND created_at > $2`,
		contentHash, time.Now().Add(-72*time.Hour)).Scan(&count)
	return count, err
}

func (s *FraudContentStore) LogContentCheck(ctx context.Context, check *ContentCheck) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO content_checks (id, task_id, creator_id, content_type, content_hash,
		 risk_score, flags, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		check.ID, check.TaskID, check.CreatorID, check.ContentType, check.ContentHash,
		check.RiskScore, check.Flags, check.Status, check.CreatedAt)
	return err
}

func (s *FraudContentStore) ListContentChecks(ctx context.Context, status string, limit, offset int) ([]ContentCheck, error) {
	var rows *sql.Rows
	var err error
	if status != "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, task_id, creator_id, content_type, content_hash,
			 risk_score, flags, status, created_at
			 FROM content_checks WHERE status = $1
			 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			status, limit, offset)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, task_id, creator_id, content_type, content_hash,
			 risk_score, flags, status, created_at
			 FROM content_checks ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []ContentCheck
	for rows.Next() {
		var check ContentCheck
		if err := rows.Scan(&check.ID, &check.TaskID, &check.CreatorID, &check.ContentType,
			&check.ContentHash, &check.RiskScore, &check.Flags, &check.Status, &check.CreatedAt); err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	return checks, rows.Err()
}

type ContentCheck struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	CreatorID   string    `json:"creator_id"`
	ContentType string    `json:"content_type"`
	ContentHash string    `json:"content_hash"`
	RiskScore   float64   `json:"risk_score"`
	Flags       string    `json:"flags"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type FraudContentHandler struct {
	store *FraudContentStore
}

func NewFraudContentHandler(store *FraudContentStore) *FraudContentHandler {
	return &FraudContentHandler{store: store}
}

func (h *FraudContentHandler) HandleCheckContent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TaskID      string `json:"task_id"`
		CreatorID   string `json:"creator_id"`
		ContentType string `json:"content_type"`
		ContentHash string `json:"content_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if req.ContentHash == "" {
		jsonError(w, 400, "content_hash required")
		return
	}

	dups, err := h.store.CheckDuplicateContent(r.Context(), req.ContentHash)
	if err != nil {
		jsonError(w, 500, "check content: "+err.Error())
		return
	}

	riskScore := 0.0
	flags := ""
	status := "clean"

	if dups >= 3 {
		riskScore = 0.8
		flags = "duplicate_content"
		status = "suspicious"
	} else if dups >= 1 {
		riskScore = 0.3
		flags = "recent_duplicate"
		status = "review"
	}

	check := &ContentCheck{
		ID:          GenerateID(),
		TaskID:      req.TaskID,
		CreatorID:   req.CreatorID,
		ContentType: req.ContentType,
		ContentHash: req.ContentHash,
		RiskScore:   riskScore,
		Flags:       flags,
		Status:      status,
		CreatedAt:   time.Now(),
	}

	if err := h.store.LogContentCheck(r.Context(), check); err != nil {
		jsonError(w, 500, "log content check: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]interface{}{
		"check":      check,
		"duplicates": dups,
		"is_safe":    status == "clean",
	})
}

func (h *FraudContentHandler) HandleListChecks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	if limit > 100 {
		limit = 100
	}
	offset := parseInt(r.URL.Query().Get("offset"), 0)

	checks, err := h.store.ListContentChecks(r.Context(), status, limit, offset)
	if err != nil {
		jsonError(w, 500, "list content checks: "+err.Error())
		return
	}
	if checks == nil {
		checks = []ContentCheck{}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"content_checks": checks,
		"total":          len(checks),
	})
}
