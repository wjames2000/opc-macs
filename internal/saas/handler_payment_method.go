package saas

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type PaymentMethodStore struct {
	db *sql.DB
}

func NewPaymentMethodStore(db *sql.DB) *PaymentMethodStore {
	return &PaymentMethodStore{db: db}
}

func (s *PaymentMethodStore) Upsert(ctx context.Context, pm *PaymentMethod) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO payment_methods (id, tenant_id, user_id, platform, account_info, is_default, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (user_id, platform) DO UPDATE SET
		   account_info = $5, is_default = $6, updated_at = $8`,
		pm.ID, pm.TenantID, pm.UserID, pm.Platform, pm.AccountInfo, pm.IsDefault,
		pm.CreatedAt, pm.UpdatedAt)
	return err
}

func (s *PaymentMethodStore) ListByUser(ctx context.Context, userID string) ([]PaymentMethod, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, user_id, platform, account_info, is_default, created_at, updated_at
		 FROM payment_methods WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pms []PaymentMethod
	for rows.Next() {
		var pm PaymentMethod
		if err := rows.Scan(&pm.ID, &pm.TenantID, &pm.UserID, &pm.Platform, &pm.AccountInfo,
			&pm.IsDefault, &pm.CreatedAt, &pm.UpdatedAt); err != nil {
			return nil, err
		}
		pms = append(pms, pm)
	}
	return pms, rows.Err()
}

func (s *PaymentMethodStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM payment_methods WHERE id = $1`, id)
	return err
}

func (s *PaymentMethodStore) SetDefault(ctx context.Context, userID, id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE payment_methods SET is_default = false WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE payment_methods SET is_default = true, updated_at = $1 WHERE id = $2 AND user_id = $3`,
		time.Now(), id, userID); err != nil {
		return err
	}

	return tx.Commit()
}

type PaymentMethod struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id"`
	Platform    string    `json:"platform"`
	AccountInfo string    `json:"account_info"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PaymentMethodHandler struct {
	store *PaymentMethodStore
}

func NewPaymentMethodHandler(store *PaymentMethodStore) *PaymentMethodHandler {
	return &PaymentMethodHandler{store: store}
}

func (h *PaymentMethodHandler) HandleUpsert(w http.ResponseWriter, r *http.Request) {
	tenantID, userID := getTenantID(r), getUserID(r)

	var req struct {
		Platform    string `json:"platform"`
		AccountInfo string `json:"account_info"`
		IsDefault   bool   `json:"is_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if req.Platform == "" {
		jsonError(w, 400, "platform is required")
		return
	}
	if req.AccountInfo == "" {
		jsonError(w, 400, "account_info is required")
		return
	}

	pm := &PaymentMethod{
		ID:          GenerateID(),
		TenantID:    tenantID,
		UserID:      userID,
		Platform:    req.Platform,
		AccountInfo: req.AccountInfo,
		IsDefault:   req.IsDefault,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.store.Upsert(r.Context(), pm); err != nil {
		jsonError(w, 500, "save payment method: "+err.Error())
		return
	}

	jsonResponse(w, 200, pm)
}

func (h *PaymentMethodHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	_, userID := getTenantID(r), getUserID(r)

	pms, err := h.store.ListByUser(r.Context(), userID)
	if err != nil {
		jsonError(w, 500, "list payment methods: "+err.Error())
		return
	}
	if pms == nil {
		pms = []PaymentMethod{}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"payment_methods": pms,
		"total":           len(pms),
	})
}

func (h *PaymentMethodHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "payment method id required")
		return
	}

	if err := h.store.Delete(r.Context(), id); err != nil {
		jsonError(w, 500, "delete payment method: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]string{"status": "deleted"})
}

func (h *PaymentMethodHandler) HandleSetDefault(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "payment method id required")
		return
	}
	_, userID := getTenantID(r), getUserID(r)

	if err := h.store.SetDefault(r.Context(), userID, id); err != nil {
		jsonError(w, 500, "set default: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]string{"status": "default_updated"})
}
