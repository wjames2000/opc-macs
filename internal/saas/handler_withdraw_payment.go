package saas

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type WithdrawPaymentStore struct {
	db *sql.DB
}

func NewWithdrawPaymentStore(db *sql.DB) *WithdrawPaymentStore {
	return &WithdrawPaymentStore{db: db}
}

func (s *WithdrawPaymentStore) Create(ctx context.Context, p *WithdrawPayment) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO withdraw_payments (id, tenant_id, vendor_id, creator_user_id, amount,
		 platform, account_info, status, notes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		p.ID, p.TenantID, p.VendorID, p.CreatorUserID, p.Amount,
		p.Platform, p.AccountInfo, p.Status, p.Notes, p.CreatedAt, p.UpdatedAt)
	return err
}

func (s *WithdrawPaymentStore) GetByID(ctx context.Context, id string) (*WithdrawPayment, error) {
	p := &WithdrawPayment{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, vendor_id, creator_user_id, amount, platform, account_info,
		 status, notes, created_at, updated_at FROM withdraw_payments WHERE id = $1`, id).
		Scan(&p.ID, &p.TenantID, &p.VendorID, &p.CreatorUserID, &p.Amount,
			&p.Platform, &p.AccountInfo, &p.Status, &p.Notes, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (s *WithdrawPaymentStore) ListByVendor(ctx context.Context, vendorID string, limit, offset int) ([]WithdrawPayment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, vendor_id, creator_user_id, amount, platform, account_info,
		 status, notes, created_at, updated_at FROM withdraw_payments
		 WHERE vendor_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		vendorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ps []WithdrawPayment
	for rows.Next() {
		var p WithdrawPayment
		if err := rows.Scan(&p.ID, &p.TenantID, &p.VendorID, &p.CreatorUserID, &p.Amount,
			&p.Platform, &p.AccountInfo, &p.Status, &p.Notes, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	return ps, rows.Err()
}

func (s *WithdrawPaymentStore) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE withdraw_payments SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), id)
	return err
}

type WithdrawPayment struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	VendorID      string    `json:"vendor_id"`
	CreatorUserID string    `json:"creator_user_id"`
	Amount        float64   `json:"amount"`
	Platform      string    `json:"platform"`
	AccountInfo   string    `json:"account_info"`
	Status        string    `json:"status"`
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type WithdrawPaymentHandler struct {
	store *WithdrawPaymentStore
}

func NewWithdrawPaymentHandler(store *WithdrawPaymentStore) *WithdrawPaymentHandler {
	return &WithdrawPaymentHandler{store: store}
}

func (h *WithdrawPaymentHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID, userID := getTenantID(r), getUserID(r)

	var req struct {
		CreatorUserID string  `json:"creator_user_id"`
		Amount        float64 `json:"amount"`
		Platform      string  `json:"platform"`
		AccountInfo   string  `json:"account_info"`
		Notes         string  `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if req.CreatorUserID == "" {
		jsonError(w, 400, "creator_user_id is required")
		return
	}
	if req.Amount <= 0 {
		jsonError(w, 400, "amount must be positive")
		return
	}
	if req.Platform == "" {
		jsonError(w, 400, "platform is required")
		return
	}

	p := &WithdrawPayment{
		ID:            GenerateID(),
		TenantID:      tenantID,
		VendorID:      userID,
		CreatorUserID: req.CreatorUserID,
		Amount:        req.Amount,
		Platform:      req.Platform,
		AccountInfo:   req.AccountInfo,
		Status:        "pending",
		Notes:         req.Notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.store.Create(r.Context(), p); err != nil {
		jsonError(w, 500, "create payment: "+err.Error())
		return
	}

	jsonResponse(w, 201, p)
}

func (h *WithdrawPaymentHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	_, userID := getTenantID(r), getUserID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	ps, err := h.store.ListByVendor(r.Context(), userID, limit, offset)
	if err != nil {
		jsonError(w, 500, "list payments: "+err.Error())
		return
	}
	if ps == nil {
		ps = []WithdrawPayment{}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"payments": ps,
		"total":    len(ps),
	})
}

func (h *WithdrawPaymentHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "payment id required")
		return
	}

	p, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, 500, "get payment: "+err.Error())
		return
	}
	if p == nil {
		jsonError(w, 404, "payment not found")
		return
	}

	jsonResponse(w, 200, p)
}

func (h *WithdrawPaymentHandler) HandleProcess(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "payment id required")
		return
	}

	var req struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}

	var newStatus string
	switch req.Action {
	case "approve":
		newStatus = "approved"
	case "reject":
		newStatus = "rejected"
	default:
		jsonError(w, 400, "action must be 'approve' or 'reject'")
		return
	}

	if err := h.store.UpdateStatus(r.Context(), id, newStatus); err != nil {
		jsonError(w, 500, "process payment: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]string{"status": newStatus})
}
