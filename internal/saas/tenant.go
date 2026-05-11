package saas

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type TenantStore struct {
	db *sql.DB
}

func NewTenantStore(db *sql.DB) *TenantStore {
	return &TenantStore{db: db}
}

func (s *TenantStore) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS tenants (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		slug TEXT UNIQUE NOT NULL,
		plan TEXT NOT NULL DEFAULT 'free',
		status TEXT NOT NULL DEFAULT 'active',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		email TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'member',
		status TEXT NOT NULL DEFAULT 'active',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(tenant_id, email)
	);
	CREATE TABLE IF NOT EXISTS tenant_models (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		provider TEXT NOT NULL,
		model_name TEXT NOT NULL,
		api_key TEXT NOT NULL DEFAULT '',
		api_base_url TEXT NOT NULL DEFAULT '',
		is_default BOOLEAN DEFAULT false,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE TABLE IF NOT EXISTS token_usage (
		id BIGSERIAL PRIMARY KEY,
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		user_id TEXT NOT NULL DEFAULT '',
		agent_name TEXT NOT NULL,
		model_name TEXT NOT NULL,
		input_tokens INTEGER NOT NULL DEFAULT 0,
		output_tokens INTEGER NOT NULL DEFAULT 0,
		cost DECIMAL(12,6) NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_token_usage_tenant ON token_usage(tenant_id, created_at);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *TenantStore) Create(ctx context.Context, name, slug string) (*Tenant, error) {
	now := time.Now()
	t := &Tenant{
		ID:        GenerateID(),
		Name:      name,
		Slug:      slug,
		Plan:      "free",
		Status:    "active",
		CreatedAt: now,
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO tenants (id, name, slug, plan, status, created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		t.ID, t.Name, t.Slug, t.Plan, t.Status, t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("saas: create tenant: %w", err)
	}
	return t, nil
}

func (s *TenantStore) GetByID(ctx context.Context, id string) (*Tenant, error) {
	t := &Tenant{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, slug, plan, status, created_at FROM tenants WHERE id=$1`, id).
		Scan(&t.ID, &t.Name, &t.Slug, &t.Plan, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("saas: get tenant: %w", err)
	}
	return t, nil
}

func (s *TenantStore) GetByAPIKey(ctx context.Context, apiKey string) (*Tenant, error) {
	t := &Tenant{}
	err := s.db.QueryRowContext(ctx,
		`SELECT t.id, t.name, t.slug, t.plan, t.status, t.created_at
		 FROM tenants t JOIN users u ON u.tenant_id = t.id
		 WHERE u.id = $1`, apiKey).Scan(&t.ID, &t.Name, &t.Slug, &t.Plan, &t.Status, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("saas: get tenant by api key: %w", err)
	}
	return t, nil
}

func (s *TenantStore) GetAll(ctx context.Context) ([]*Tenant, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, slug, plan, status, created_at FROM tenants ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tenants []*Tenant
	for rows.Next() {
		t := &Tenant{}
		rows.Scan(&t.ID, &t.Name, &t.Slug, &t.Plan, &t.Status, &t.CreatedAt)
		tenants = append(tenants, t)
	}
	return tenants, nil
}

func (s *TenantStore) UpdatePlan(ctx context.Context, id, plan string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE tenants SET plan=$1 WHERE id=$2`, plan, id)
	return err
}
