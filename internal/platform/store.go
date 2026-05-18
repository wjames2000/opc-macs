package platform

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type DBTokenStore struct {
	db *sql.DB
}

func NewDBTokenStore(db *sql.DB) *DBTokenStore {
	return &DBTokenStore{db: db}
}

func (s *DBTokenStore) Get(ctx context.Context, accountID string) (*AuthToken, error) {
	var tokenJSON string
	err := s.db.QueryRowContext(ctx,
		`SELECT auth_token FROM platform_accounts WHERE id=$1 AND status='active'`, accountID).Scan(&tokenJSON)
	if err != nil {
		return nil, fmt.Errorf("token store: get: %w", err)
	}

	var token AuthToken
	if err := json.Unmarshal([]byte(tokenJSON), &token); err != nil {
		return nil, fmt.Errorf("token store: unmarshal: %w", err)
	}
	token.Raw = map[string]any{"account_id": accountID}
	return &token, nil
}

func (s *DBTokenStore) Save(ctx context.Context, accountID string, token *AuthToken) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("token store: marshal: %w", err)
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE platform_accounts SET auth_token=$1, token_expires_at=$2, updated_at=NOW() WHERE id=$3`,
		string(data), token.ExpiresAt, accountID)
	if err != nil {
		return fmt.Errorf("token store: save: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("token store: account %s not found", accountID)
	}
	return nil
}

func (s *DBTokenStore) Delete(ctx context.Context, accountID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE platform_accounts SET status='revoked', auth_token='', updated_at=NOW() WHERE id=$1`, accountID)
	if err != nil {
		return fmt.Errorf("token store: delete: %w", err)
	}
	return nil
}

func (s *DBTokenStore) List(ctx context.Context) ([]TokenEntry, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, platform_name, platform_user_id, status, auth_token, token_expires_at
		FROM platform_accounts WHERE status='active'`)
	if err != nil {
		return nil, fmt.Errorf("token store: list: %w", err)
	}
	defer rows.Close()

	var entries []TokenEntry
	for rows.Next() {
		var e TokenEntry
		var tokenJSON string
		if err := rows.Scan(&e.ID, &e.Platform, &e.PlatformUserID, &e.Status, &tokenJSON, &e.TokenExpiresAt); err != nil {
			return nil, fmt.Errorf("token store: scan: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

type TokenEntry struct {
	ID             string
	Platform       string
	PlatformUserID string
	TokenExpiresAt time.Time
	Status         string
}
