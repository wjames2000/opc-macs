package saas

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, tenantID, email, password string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("saas: hash password: %w", err)
	}
	now := time.Now()
	u := &User{
		ID:           GenerateID(),
		TenantID:     tenantID,
		Email:        email,
		PasswordHash: string(hash),
		Role:         "owner",
		Status:       "active",
		CreatedAt:    now,
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO users (id, tenant_id, email, password_hash, role, status, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		u.ID, u.TenantID, u.Email, u.PasswordHash, u.Role, u.Status, u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("saas: create user: %w", err)
	}
	return u, nil
}

func (s *UserStore) Authenticate(ctx context.Context, email, password string) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, email, password_hash, role, status, created_at FROM users WHERE email=$1`,
		email).Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Role, &u.Status, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("saas: user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("saas: invalid password")
	}
	return u, nil
}

func (s *UserStore) GetByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, email, '', role, status, created_at FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.TenantID, &u.Email, &u.Role, &u.Status, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("saas: get user: %w", err)
	}
	return u, nil
}

func (s *UserStore) GetByTenant(ctx context.Context, tenantID string) ([]*User, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tenant_id, email, '', role, status, created_at FROM users WHERE tenant_id=$1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []*User
	for rows.Next() {
		u := &User{}
		rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.Role, &u.Status, &u.CreatedAt)
		users = append(users, u)
	}
	return users, nil
}

// GenerateToken creates a simple session token (SHA256 of user+time)
func GenerateToken(userID string) string {
	h := sha256.Sum256([]byte(userID + time.Now().String()))
	return fmt.Sprintf("%x", h)
}
