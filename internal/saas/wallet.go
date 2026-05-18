package saas

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Wallet struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id"`
	Balance     float64   `json:"balance"`
	Frozen      float64   `json:"frozen"`
	TotalEarned float64   `json:"total_earned"`
	TotalSpent  float64   `json:"total_spent"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WalletTransaction struct {
	ID           string    `json:"id"`
	WalletID     string    `json:"wallet_id"`
	Type         string    `json:"type"` // deposit, settlement, withdraw, freeze, unfreeze, commission
	Amount       float64   `json:"amount"`
	BalanceAfter float64   `json:"balance_after"`
	ReferenceID  string    `json:"reference_id,omitempty"` // order_id or withdraw_id
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type WalletStore struct {
	db *sql.DB
}

func NewWalletStore(db *sql.DB) *WalletStore {
	return &WalletStore{db: db}
}

func (s *WalletStore) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS wallets (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		user_id TEXT NOT NULL REFERENCES users(id),
		balance DECIMAL(14,2) NOT NULL DEFAULT 0,
		frozen DECIMAL(14,2) NOT NULL DEFAULT 0,
		total_earned DECIMAL(14,2) NOT NULL DEFAULT 0,
		total_spent DECIMAL(14,2) NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'active',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(tenant_id, user_id)
	);
	CREATE TABLE IF NOT EXISTS wallet_transactions (
		id TEXT PRIMARY KEY,
		wallet_id TEXT NOT NULL REFERENCES wallets(id),
		type TEXT NOT NULL,
		amount DECIMAL(14,2) NOT NULL DEFAULT 0,
		balance_after DECIMAL(14,2) NOT NULL DEFAULT 0,
		reference_id TEXT NOT NULL DEFAULT '',
		description TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'completed',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet ON wallet_transactions(wallet_id, created_at);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *WalletStore) EnsureWallet(ctx context.Context, tenantID, userID string) (*Wallet, error) {
	w, err := s.GetByUser(ctx, tenantID, userID)
	if err == nil {
		return w, nil
	}

	w = &Wallet{
		ID:        GenerateID(),
		TenantID:  tenantID,
		UserID:    userID,
		Balance:   0,
		Frozen:    0,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO wallets (id, tenant_id, user_id, balance, frozen, total_earned, total_spent, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		w.ID, w.TenantID, w.UserID, w.Balance, w.Frozen, w.TotalEarned, w.TotalSpent, w.Status, w.CreatedAt, w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("wallet: ensure wallet: %w", err)
	}
	return w, nil
}

func (s *WalletStore) GetByUser(ctx context.Context, tenantID, userID string) (*Wallet, error) {
	w := &Wallet{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, user_id, balance, frozen, total_earned, total_spent, status, created_at, updated_at
		FROM wallets WHERE tenant_id=$1 AND user_id=$2`, tenantID, userID).
		Scan(&w.ID, &w.TenantID, &w.UserID, &w.Balance, &w.Frozen, &w.TotalEarned, &w.TotalSpent, &w.Status, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("wallet: get by user: %w", err)
	}
	return w, nil
}

func (s *WalletStore) Deposit(ctx context.Context, walletID string, amount float64, description string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("wallet: deposit begin tx: %w", err)
	}
	defer tx.Rollback()

	var balance float64
	err = tx.QueryRowContext(ctx,
		`UPDATE wallets SET balance = balance + $1, total_spent = total_spent + $2, updated_at = NOW()
		WHERE id=$3 RETURNING balance`, amount, amount, walletID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("wallet: deposit update: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO wallet_transactions (id, wallet_id, type, amount, balance_after, description, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		GenerateID(), walletID, "deposit", amount, balance, description, "completed", time.Now())
	if err != nil {
		return fmt.Errorf("wallet: deposit insert tx: %w", err)
	}

	return tx.Commit()
}

func (s *WalletStore) Freeze(ctx context.Context, walletID string, amount float64, referenceID, description string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("wallet: freeze begin tx: %w", err)
	}
	defer tx.Rollback()

	var balance, frozen float64
	err = tx.QueryRowContext(ctx,
		`UPDATE wallets SET balance = balance - $1, frozen = frozen + $1, updated_at = NOW()
		WHERE id=$2 AND balance >= $1 RETURNING balance, frozen`,
		amount, walletID).Scan(&balance, &frozen)
	if err != nil {
		return fmt.Errorf("wallet: freeze update (insufficient balance?): %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO wallet_transactions (id, wallet_id, type, amount, balance_after, reference_id, description, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		GenerateID(), walletID, "freeze", amount, balance, referenceID, description, "completed", time.Now())
	if err != nil {
		return fmt.Errorf("wallet: freeze insert tx: %w", err)
	}

	return tx.Commit()
}

func (s *WalletStore) Unfreeze(ctx context.Context, walletID string, amount float64, referenceID, description string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE wallets SET balance = balance + $1, frozen = frozen - $1, updated_at = NOW()
		WHERE id=$2 AND frozen >= $1`, amount, walletID)
	if err != nil {
		return fmt.Errorf("wallet: unfreeze: %w", err)
	}
	return nil
}

func (s *WalletStore) Settle(ctx context.Context, walletID string, amount float64, referenceID, description string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("wallet: settle begin tx: %w", err)
	}
	defer tx.Rollback()

	var balance float64
	err = tx.QueryRowContext(ctx,
		`UPDATE wallets SET balance = balance + $1, frozen = frozen - $1, total_earned = total_earned + $1, updated_at = NOW()
		WHERE id=$2 RETURNING balance`, amount, amount, walletID).Scan(&balance)
	if err != nil {
		return fmt.Errorf("wallet: settle update: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO wallet_transactions (id, wallet_id, type, amount, balance_after, reference_id, description, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		GenerateID(), walletID, "settlement", amount, balance, referenceID, description, "completed", time.Now())
	if err != nil {
		return fmt.Errorf("wallet: settle insert tx: %w", err)
	}

	return tx.Commit()
}

func (s *WalletStore) GetTransactions(ctx context.Context, walletID string, limit, offset int) ([]*WalletTransaction, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, wallet_id, type, amount, balance_after, reference_id, description, status, created_at
		FROM wallet_transactions WHERE wallet_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		walletID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("wallet: get transactions: %w", err)
	}
	defer rows.Close()

	var txs []*WalletTransaction
	for rows.Next() {
		t := &WalletTransaction{}
		if err := rows.Scan(&t.ID, &t.WalletID, &t.Type, &t.Amount, &t.BalanceAfter, &t.ReferenceID, &t.Description, &t.Status, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("wallet: scan tx: %w", err)
		}
		txs = append(txs, t)
	}
	return txs, nil
}
