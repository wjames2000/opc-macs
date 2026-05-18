package saas

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PlatformAccount struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	UserID           string     `json:"user_id"`
	PlatformName     string     `json:"platform_name"`
	PlatformUserID   string     `json:"platform_user_id"`
	PlatformUsername string     `json:"platform_username"`
	AvatarURL        string     `json:"avatar_url"`
	AuthToken        string     `json:"-"`
	TokenExpiresAt   *time.Time `json:"token_expires_at,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type PlatformAccountStore struct {
	db *sql.DB
}

func NewPlatformAccountStore(db *sql.DB) *PlatformAccountStore {
	return &PlatformAccountStore{db: db}
}

func (s *PlatformAccountStore) InitSchema() error {
	_, err := s.db.Exec(`
	CREATE TABLE IF NOT EXISTS platform_accounts (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		user_id TEXT NOT NULL REFERENCES users(id),
		platform_name TEXT NOT NULL,
		platform_user_id TEXT NOT NULL DEFAULT '',
		platform_username TEXT NOT NULL DEFAULT '',
		avatar_url TEXT NOT NULL DEFAULT '',
		auth_token TEXT NOT NULL DEFAULT '',
		token_expires_at TIMESTAMPTZ,
		status TEXT NOT NULL DEFAULT 'active',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(tenant_id, platform_name, platform_user_id)
	);
	CREATE INDEX IF NOT EXISTS idx_platform_accounts_tenant_plaf
		ON platform_accounts(tenant_id, platform_name);
	`)
	return err
}

func (s *PlatformAccountStore) Create(ctx context.Context, a *PlatformAccount) error {
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO platform_accounts (id, tenant_id, user_id, platform_name, platform_user_id, platform_username, avatar_url, auth_token, token_expires_at, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		a.ID, a.TenantID, a.UserID, a.PlatformName, a.PlatformUserID, a.PlatformUsername, a.AvatarURL, a.AuthToken, a.TokenExpiresAt, a.Status, a.CreatedAt, a.UpdatedAt)
	return err
}

func (s *PlatformAccountStore) GetByID(ctx context.Context, id string) (*PlatformAccount, error) {
	a := &PlatformAccount{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, user_id, platform_name, platform_user_id, platform_username, avatar_url, auth_token, token_expires_at, status, created_at, updated_at
		FROM platform_accounts WHERE id=$1`, id).
		Scan(&a.ID, &a.TenantID, &a.UserID, &a.PlatformName, &a.PlatformUserID, &a.PlatformUsername, &a.AvatarURL, &a.AuthToken, &a.TokenExpiresAt, &a.Status, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("platform account: get by id: %w", err)
	}
	return a, nil
}

func (s *PlatformAccountStore) GetByPlatform(ctx context.Context, tenantID, platformName string) ([]*PlatformAccount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, user_id, platform_name, platform_user_id, platform_username, avatar_url, auth_token, token_expires_at, status, created_at, updated_at
		FROM platform_accounts WHERE tenant_id=$1 AND platform_name=$2 ORDER BY created_at DESC`,
		tenantID, platformName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAccounts(rows)
}

func (s *PlatformAccountStore) ListByTenant(ctx context.Context, tenantID string) ([]*PlatformAccount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, user_id, platform_name, platform_user_id, platform_username, avatar_url, auth_token, token_expires_at, status, created_at, updated_at
		FROM platform_accounts WHERE tenant_id=$1 ORDER BY platform_name, created_at DESC`,
		tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAccounts(rows)
}

func (s *PlatformAccountStore) UpdateToken(ctx context.Context, id, tokenJSON string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE platform_accounts SET auth_token=$1, token_expires_at=$2, updated_at=NOW() WHERE id=$3`,
		tokenJSON, expiresAt, id)
	return err
}

func (s *PlatformAccountStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE platform_accounts SET status='revoked', updated_at=NOW() WHERE id=$1`, id)
	return err
}

func scanAccounts(rows *sql.Rows) ([]*PlatformAccount, error) {
	var accounts []*PlatformAccount
	for rows.Next() {
		a := &PlatformAccount{}
		if err := rows.Scan(&a.ID, &a.TenantID, &a.UserID, &a.PlatformName, &a.PlatformUserID, &a.PlatformUsername, &a.AvatarURL, &a.AuthToken, &a.TokenExpiresAt, &a.Status, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}
