package saas

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Notification struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"` // order_status, withdrawal, review, system
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	Link      string    `json:"link,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationStore struct {
	db *sql.DB
}

func NewNotificationStore(db *sql.DB) *NotificationStore {
	return &NotificationStore{db: db}
}

func (s *NotificationStore) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL,
		user_id TEXT NOT NULL,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		is_read BOOLEAN DEFAULT FALSE,
		link TEXT,
		created_at TIMESTAMP DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, is_read);
	CREATE INDEX IF NOT EXISTS idx_notifications_tenant ON notifications(tenant_id, created_at);`
	_, err := s.db.Exec(schema)
	return err
}

func (s *NotificationStore) Create(ctx context.Context, n *Notification) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO notifications (id, tenant_id, user_id, type, title, content, link, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		n.ID, n.TenantID, n.UserID, n.Type, n.Title, n.Content, n.Link, n.CreatedAt)
	return err
}

func (s *NotificationStore) ListByUser(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]*Notification, error) {
	query := `SELECT id, tenant_id, user_id, type, title, content, is_read, link, created_at
		FROM notifications WHERE user_id=$1`
	args := []interface{}{userID}
	argIdx := 2
	if unreadOnly {
		query += fmt.Sprintf(" AND is_read=FALSE")
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
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ns []*Notification
	for rows.Next() {
		n := &Notification{}
		if err := rows.Scan(&n.ID, &n.TenantID, &n.UserID, &n.Type, &n.Title, &n.Content, &n.IsRead, &n.Link, &n.CreatedAt); err != nil {
			return nil, err
		}
		ns = append(ns, n)
	}
	return ns, nil
}

func (s *NotificationStore) MarkRead(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE notifications SET is_read=TRUE WHERE id=$1`, id)
	return err
}

func (s *NotificationStore) MarkAllRead(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE notifications SET is_read=TRUE WHERE user_id=$1`, userID)
	return err
}

func (s *NotificationStore) CountUnread(ctx context.Context, userID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id=$1 AND is_read=FALSE`, userID).Scan(&count)
	return count, err
}

func (s *NotificationStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM notifications WHERE id=$1`, id)
	return err
}
