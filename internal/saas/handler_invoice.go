package saas

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type InvoiceStore struct {
	db *sql.DB
}

func NewInvoiceStore(db *sql.DB) *InvoiceStore {
	return &InvoiceStore{db: db}
}

func (s *InvoiceStore) Create(ctx context.Context, inv *Invoice) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO invoices (id, tenant_id, advertiser_id, amount, tax_amount, total_amount,
		 status, platform, period_start, period_end, notes, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		inv.ID, inv.TenantID, inv.AdvertiserID, inv.Amount, inv.TaxAmount, inv.TotalAmount,
		inv.Status, inv.Platform, inv.PeriodStart, inv.PeriodEnd, inv.Notes,
		inv.CreatedAt, inv.UpdatedAt)
	return err
}

func (s *InvoiceStore) GetByID(ctx context.Context, id string) (*Invoice, error) {
	inv := &Invoice{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, advertiser_id, amount, tax_amount, total_amount,
		 status, platform, period_start, period_end, notes, created_at, updated_at
		 FROM invoices WHERE id = $1`, id).
		Scan(&inv.ID, &inv.TenantID, &inv.AdvertiserID, &inv.Amount, &inv.TaxAmount,
			&inv.TotalAmount, &inv.Status, &inv.Platform, &inv.PeriodStart, &inv.PeriodEnd,
			&inv.Notes, &inv.CreatedAt, &inv.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return inv, err
}

func (s *InvoiceStore) ListByAdvertiser(ctx context.Context, advertiserID string, limit, offset int) ([]Invoice, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, advertiser_id, amount, tax_amount, total_amount,
		 status, platform, period_start, period_end, notes, created_at, updated_at
		 FROM invoices WHERE advertiser_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		advertiserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invs []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(&inv.ID, &inv.TenantID, &inv.AdvertiserID, &inv.Amount, &inv.TaxAmount,
			&inv.TotalAmount, &inv.Status, &inv.Platform, &inv.PeriodStart, &inv.PeriodEnd,
			&inv.Notes, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}
	return invs, rows.Err()
}

func (s *InvoiceStore) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE invoices SET status = $1, updated_at = $2 WHERE id = $3`,
		status, time.Now(), id)
	return err
}

type Invoice struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	AdvertiserID string    `json:"advertiser_id"`
	Amount       float64   `json:"amount"`
	TaxAmount    float64   `json:"tax_amount"`
	TotalAmount  float64   `json:"total_amount"`
	Status       string    `json:"status"`
	Platform     string    `json:"platform"`
	PeriodStart  string    `json:"period_start"`
	PeriodEnd    string    `json:"period_end"`
	Notes        string    `json:"notes,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type InvoiceHandler struct {
	store *InvoiceStore
}

func NewInvoiceHandler(store *InvoiceStore) *InvoiceHandler {
	return &InvoiceHandler{store: store}
}

func (h *InvoiceHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	tenantID, userID := getTenantID(r), getUserID(r)

	var req struct {
		Amount      float64 `json:"amount"`
		TaxRate     float64 `json:"tax_rate"`
		Platform    string  `json:"platform"`
		PeriodStart string  `json:"period_start"`
		PeriodEnd   string  `json:"period_end"`
		Notes       string  `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if req.Amount <= 0 {
		jsonError(w, 400, "amount must be positive")
		return
	}
	if req.TaxRate < 0 {
		jsonError(w, 400, "tax_rate must be non-negative")
		return
	}

	inv := &Invoice{
		ID:           GenerateID(),
		TenantID:     tenantID,
		AdvertiserID: userID,
		Amount:       req.Amount,
		TaxAmount:    req.Amount * req.TaxRate / 100.0,
		TotalAmount:  req.Amount + req.Amount*req.TaxRate/100.0,
		Status:       "pending",
		Platform:     req.Platform,
		PeriodStart:  req.PeriodStart,
		PeriodEnd:    req.PeriodEnd,
		Notes:        req.Notes,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.store.Create(r.Context(), inv); err != nil {
		jsonError(w, 500, "create invoice: "+err.Error())
		return
	}

	jsonResponse(w, 201, inv)
}

func (h *InvoiceHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	_, userID := getTenantID(r), getUserID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	invs, err := h.store.ListByAdvertiser(r.Context(), userID, limit, offset)
	if err != nil {
		jsonError(w, 500, "list invoices: "+err.Error())
		return
	}
	if invs == nil {
		invs = []Invoice{}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"invoices": invs,
		"total":    len(invs),
	})
}

func (h *InvoiceHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "invoice id required")
		return
	}

	inv, err := h.store.GetByID(r.Context(), id)
	if err != nil {
		jsonError(w, 500, "get invoice: "+err.Error())
		return
	}
	if inv == nil {
		jsonError(w, 404, "invoice not found")
		return
	}

	jsonResponse(w, 200, inv)
}

func (h *InvoiceHandler) HandlePay(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "invoice id required")
		return
	}

	if err := h.store.UpdateStatus(r.Context(), id, "paid"); err != nil {
		jsonError(w, 500, "pay invoice: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]string{"status": "paid"})
}

func (h *InvoiceHandler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "invoice id required")
		return
	}

	if err := h.store.UpdateStatus(r.Context(), id, "cancelled"); err != nil {
		jsonError(w, 500, "cancel invoice: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]string{"status": "cancelled"})
}
