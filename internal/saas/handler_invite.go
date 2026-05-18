package saas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Invitation struct {
	ID          string     `json:"id"`
	TaskID      string     `json:"task_id"`
	FromUserID  string     `json:"from_user_id"`
	ToUserID    string     `json:"to_user_id"`
	Status      string     `json:"status"`
	Message     string     `json:"message,omitempty"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type InviteHandler struct {
	store *MarketplaceStore
}

func NewInviteHandler(store *MarketplaceStore) *InviteHandler {
	return &InviteHandler{store: store}
}

func (h *InviteHandler) HandleInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}
	fromUserID := getUserID(r)

	var req struct {
		TaskID  string `json:"task_id"`
		ToUser  string `json:"to_user"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if req.TaskID == "" || req.ToUser == "" {
		jsonError(w, 400, "task_id and to_user required")
		return
	}

	inv := &Invitation{
		ID:         GenerateID(),
		TaskID:     req.TaskID,
		FromUserID: fromUserID,
		ToUserID:   req.ToUser,
		Status:     "pending",
		Message:    req.Message,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := h.store.db.QueryRowContext(r.Context(),
		`INSERT INTO invitations (id, task_id, from_user_id, to_user_id, status, message, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		inv.ID, inv.TaskID, inv.FromUserID, inv.ToUserID, inv.Status, inv.Message, inv.CreatedAt, inv.UpdatedAt,
	).Scan(&inv.ID); err != nil {
		jsonError(w, 500, "create invite: "+err.Error())
		return
	}

	jsonResponse(w, 201, inv)
}

func (h *InviteHandler) HandleListInvitations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonError(w, 405, "method not allowed")
		return
	}
	userID := getUserID(r)
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	query := `SELECT id, task_id, from_user_id, to_user_id, status, message, responded_at, created_at, updated_at
		FROM invitations WHERE to_user_id=$1`
	args := []interface{}{userID}
	argIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND status=$%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, limit)
		argIdx++
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, offset)
	}

	rows, err := h.store.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		jsonError(w, 500, "list invitations: "+err.Error())
		return
	}
	defer rows.Close()

	var invs []*Invitation
	for rows.Next() {
		inv := &Invitation{}
		if err := rows.Scan(&inv.ID, &inv.TaskID, &inv.FromUserID, &inv.ToUserID,
			&inv.Status, &inv.Message, &inv.RespondedAt, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			jsonError(w, 500, "scan: "+err.Error())
			return
		}
		invs = append(invs, inv)
	}
	jsonResponse(w, 200, invs)
}

func (h *InviteHandler) HandleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}
	invID := r.PathValue("id")
	if invID == "" {
		jsonError(w, 400, "invitation id required")
		return
	}

	now := time.Now()
	result, err := h.store.db.ExecContext(r.Context(),
		`UPDATE invitations SET status='accepted', responded_at=$1, updated_at=$1 WHERE id=$2 AND status='pending'`,
		now, invID)
	if err != nil {
		jsonError(w, 500, "accept: "+err.Error())
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		jsonError(w, 404, "invitation not found or already responded")
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "accepted"})
}

func (h *InviteHandler) HandleDeclineInvite(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}
	invID := r.PathValue("id")
	if invID == "" {
		jsonError(w, 400, "invitation id required")
		return
	}

	now := time.Now()
	result, err := h.store.db.ExecContext(r.Context(),
		`UPDATE invitations SET status='declined', responded_at=$1, updated_at=$1 WHERE id=$2 AND status='pending'`,
		now, invID)
	if err != nil {
		jsonError(w, 500, "decline: "+err.Error())
		return
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		jsonError(w, 404, "invitation not found or already responded")
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "declined"})
}

// InitSchema for invite handler
func (h *InviteHandler) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS invitations (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL REFERENCES marketplace_tasks(id),
		from_user_id TEXT NOT NULL,
		to_user_id TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		message TEXT NOT NULL DEFAULT '',
		responded_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_invitations_to_user ON invitations(to_user_id, status);
	CREATE INDEX IF NOT EXISTS idx_invitations_task ON invitations(task_id);
	`
	_, err := h.store.db.ExecContext(ctx, schema)
	return err
}
