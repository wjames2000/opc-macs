package saas

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type WithdrawalMethod string

const (
	WithdrawalAlipay   WithdrawalMethod = "alipay"
	WithdrawalWechat   WithdrawalMethod = "wechat"
	WithdrawalBankCard WithdrawalMethod = "bank_card"
)

type WithdrawalStatus string

const (
	WithdrawalPending    WithdrawalStatus = "pending"
	WithdrawalProcessing WithdrawalStatus = "processing"
	WithdrawalCompleted  WithdrawalStatus = "completed"
	WithdrawalRejected   WithdrawalStatus = "rejected"
	WithdrawalCancelled  WithdrawalStatus = "cancelled"
)

type Withdrawal struct {
	ID          string           `json:"id"`
	TenantID    string           `json:"tenant_id"`
	UserID      string           `json:"user_id"`
	WalletID    string           `json:"wallet_id"`
	Amount      float64          `json:"amount"`
	Fee         float64          `json:"fee"`
	NetAmount   float64          `json:"net_amount"`
	Method      WithdrawalMethod `json:"method"`
	AccountInfo string           `json:"account_info"`
	Status      WithdrawalStatus `json:"status"`
	ReviewNote  string           `json:"review_note,omitempty"`
	ProcessedBy string           `json:"processed_by,omitempty"`
	ProcessedAt *time.Time       `json:"processed_at,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

type WithdrawalStore struct {
	db *sql.DB
}

func NewWithdrawalStore(db *sql.DB) *WithdrawalStore {
	return &WithdrawalStore{db: db}
}

func (s *WithdrawalStore) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS withdrawals (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		user_id TEXT NOT NULL REFERENCES users(id),
		wallet_id TEXT NOT NULL REFERENCES wallets(id),
		amount DECIMAL(14,2) NOT NULL,
		fee DECIMAL(14,2) NOT NULL DEFAULT 0,
		net_amount DECIMAL(14,2) NOT NULL,
		method TEXT NOT NULL,
		account_info TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'pending',
		review_note TEXT NOT NULL DEFAULT '',
		processed_by TEXT NOT NULL DEFAULT '',
		processed_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_withdrawals_wallet ON withdrawals(wallet_id, created_at);
	CREATE INDEX IF NOT EXISTS idx_withdrawals_status ON withdrawals(tenant_id, status);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *WithdrawalStore) Create(ctx context.Context, w *Withdrawal) error {
	w.Status = WithdrawalPending
	w.CreatedAt = time.Now()
	w.UpdatedAt = time.Now()
	if w.Fee == 0 {
		w.Fee = w.Amount * 0.01
		if w.Fee < 1 {
			w.Fee = 1
		}
	}
	w.NetAmount = w.Amount - w.Fee

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO withdrawals (id, tenant_id, user_id, wallet_id, amount, fee, net_amount, method, account_info, status, review_note, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		w.ID, w.TenantID, w.UserID, w.WalletID, w.Amount, w.Fee, w.NetAmount, w.Method, w.AccountInfo, w.Status, "", w.CreatedAt, w.UpdatedAt)
	return err
}

func (s *WithdrawalStore) Approve(ctx context.Context, id, processedBy string) error {
	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`UPDATE withdrawals SET status='completed', processed_by=$1, processed_at=$2, updated_at=$3 WHERE id=$4 AND status='pending'`,
		processedBy, now, now, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("withdrawal %s not found or already processed", id)
	}
	return nil
}

func (s *WithdrawalStore) Reject(ctx context.Context, id, note, processedBy string) error {
	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		`UPDATE withdrawals SET status='rejected', review_note=$1, processed_by=$2, processed_at=$3, updated_at=$4 WHERE id=$5 AND status='pending'`,
		note, processedBy, now, now, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("withdrawal %s not found or already processed", id)
	}
	return nil
}

func (s *WithdrawalStore) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*Withdrawal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, user_id, wallet_id, amount, fee, net_amount, method, account_info, status, review_note, processed_by, processed_at, created_at, updated_at
		FROM withdrawals WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWithdrawals(rows)
}

func (s *WithdrawalStore) ListPending(ctx context.Context, tenantID string, limit, offset int) ([]*Withdrawal, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, user_id, wallet_id, amount, fee, net_amount, method, account_info, status, review_note, processed_by, processed_at, created_at, updated_at
		FROM withdrawals WHERE tenant_id=$1 AND status='pending' ORDER BY created_at ASC LIMIT $2 OFFSET $3`,
		tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanWithdrawals(rows)
}

func scanWithdrawals(rows *sql.Rows) ([]*Withdrawal, error) {
	var ws []*Withdrawal
	for rows.Next() {
		w := &Withdrawal{}
		if err := rows.Scan(&w.ID, &w.TenantID, &w.UserID, &w.WalletID, &w.Amount, &w.Fee, &w.NetAmount, &w.Method, &w.AccountInfo, &w.Status, &w.ReviewNote, &w.ProcessedBy, &w.ProcessedAt, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		ws = append(ws, w)
	}
	return ws, nil
}
