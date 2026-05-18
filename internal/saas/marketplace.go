package saas

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// MarketplaceTask represents a content marketing task posted by an advertiser.
type MarketplaceTask struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	CreatorID      string    `json:"creator_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Platform       string    `json:"platform"`
	ContentType    string    `json:"content_type"`
	SettlementType string    `json:"settlement_type"` // cps, cpe, cpm, hybrid, fixed
	Budget         float64   `json:"budget"`
	Requirements   string    `json:"requirements"`
	Status         string    `json:"status"` // open, in_progress, completed, cancelled
	Deadline       time.Time `json:"deadline"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TaskOrder represents an order created when a creator accepts a task.
type TaskOrder struct {
	ID               string    `json:"id"`
	TaskID           string    `json:"task_id"`
	CreatorID        string    `json:"creator_id"`
	AdvertiserID     string    `json:"advertiser_id"`
	Status           string    `json:"status"` // applied, accepted, delivered, approved, rejected, cancelled, completed
	DeliveryURL      string    `json:"delivery_url,omitempty"`
	DeliveryNote     string    `json:"delivery_note,omitempty"`
	ReviewNote       string    `json:"review_note,omitempty"`
	SettlementAmount float64   `json:"settlement_amount,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// MarketplaceStore handles marketplace task and order persistence.
type MarketplaceStore struct {
	db *sql.DB
}

func NewMarketplaceStore(db *sql.DB) *MarketplaceStore {
	return &MarketplaceStore{db: db}
}

func (s *MarketplaceStore) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS marketplace_tasks (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		creator_id TEXT NOT NULL REFERENCES users(id),
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		platform TEXT NOT NULL,
		content_type TEXT NOT NULL,
		settlement_type TEXT NOT NULL DEFAULT 'fixed',
		budget DECIMAL(12,2) NOT NULL DEFAULT 0,
		requirements TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'open',
		deadline TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_marketplace_tasks_tenant ON marketplace_tasks(tenant_id, status);
	CREATE INDEX IF NOT EXISTS idx_marketplace_tasks_platform ON marketplace_tasks(platform, status);

	CREATE TABLE IF NOT EXISTS task_orders (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL REFERENCES marketplace_tasks(id),
		creator_id TEXT NOT NULL REFERENCES users(id),
		advertiser_id TEXT NOT NULL REFERENCES users(id),
		status TEXT NOT NULL DEFAULT 'applied',
		delivery_url TEXT NOT NULL DEFAULT '',
		delivery_note TEXT NOT NULL DEFAULT '',
		review_note TEXT NOT NULL DEFAULT '',
		settlement_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(task_id, creator_id)
	);
	CREATE INDEX IF NOT EXISTS idx_task_orders_creator ON task_orders(creator_id, status);
	CREATE INDEX IF NOT EXISTS idx_task_orders_advertiser ON task_orders(advertiser_id, status);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *MarketplaceStore) CreateTask(ctx context.Context, task *MarketplaceTask) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO marketplace_tasks (id, tenant_id, creator_id, title, description, platform, content_type, settlement_type, budget, requirements, status, deadline, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		task.ID, task.TenantID, task.CreatorID, task.Title, task.Description,
		task.Platform, task.ContentType, task.SettlementType, task.Budget,
		task.Requirements, task.Status, task.Deadline, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("marketplace: create task: %w", err)
	}
	return nil
}

func (s *MarketplaceStore) GetTask(ctx context.Context, id string) (*MarketplaceTask, error) {
	t := &MarketplaceTask{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, tenant_id, creator_id, title, description, platform, content_type, settlement_type, budget, requirements, status, deadline, created_at, updated_at
		FROM marketplace_tasks WHERE id=$1`, id).
		Scan(&t.ID, &t.TenantID, &t.CreatorID, &t.Title, &t.Description,
			&t.Platform, &t.ContentType, &t.SettlementType, &t.Budget,
			&t.Requirements, &t.Status, &t.Deadline, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("marketplace: get task: %w", err)
	}
	return t, nil
}

func (s *MarketplaceStore) ListOpenTasks(ctx context.Context, platform, contentType string, limit, offset int) ([]*MarketplaceTask, error) {
	query := `SELECT id, tenant_id, creator_id, title, description, platform, content_type, settlement_type, budget, requirements, status, deadline, created_at, updated_at
		FROM marketplace_tasks WHERE status='open'`
	var args []interface{}
	argIdx := 1

	if platform != "" {
		query += fmt.Sprintf(" AND platform=$%d", argIdx)
		args = append(args, platform)
		argIdx++
	}
	if contentType != "" {
		query += fmt.Sprintf(" AND content_type=$%d", argIdx)
		args = append(args, contentType)
		argIdx++
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
		return nil, fmt.Errorf("marketplace: list open tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*MarketplaceTask
	for rows.Next() {
		t := &MarketplaceTask{}
		if err := rows.Scan(&t.ID, &t.TenantID, &t.CreatorID, &t.Title, &t.Description,
			&t.Platform, &t.ContentType, &t.SettlementType, &t.Budget,
			&t.Requirements, &t.Status, &t.Deadline, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("marketplace: scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (s *MarketplaceStore) CreateOrder(ctx context.Context, order *TaskOrder) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO task_orders (id, task_id, creator_id, advertiser_id, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		order.ID, order.TaskID, order.CreatorID, order.AdvertiserID, order.Status, order.CreatedAt, order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("marketplace: create order: %w", err)
	}
	return nil
}

func (s *MarketplaceStore) GetOrder(ctx context.Context, id string) (*TaskOrder, error) {
	o := &TaskOrder{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, task_id, creator_id, advertiser_id, status, delivery_url, delivery_note, review_note, settlement_amount, created_at, updated_at
		FROM task_orders WHERE id=$1`, id).
		Scan(&o.ID, &o.TaskID, &o.CreatorID, &o.AdvertiserID, &o.Status,
			&o.DeliveryURL, &o.DeliveryNote, &o.ReviewNote, &o.SettlementAmount,
			&o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("marketplace: get order: %w", err)
	}
	return o, nil
}

func (s *MarketplaceStore) UpdateOrderStatus(ctx context.Context, id, status string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE task_orders SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	if err != nil {
		return fmt.Errorf("marketplace: update order status: %w", err)
	}
	return nil
}

func (s *MarketplaceStore) UpdateOrderSettlement(ctx context.Context, id string, amount float64) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE task_orders SET settlement_amount=$1, updated_at=NOW() WHERE id=$2`, amount, id)
	if err != nil {
		return fmt.Errorf("marketplace: update settlement: %w", err)
	}
	return nil
}

func (s *MarketplaceStore) UpdateOrderNote(ctx context.Context, id, note string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE task_orders SET review_note=$1, updated_at=NOW() WHERE id=$2`, note, id)
	if err != nil {
		return fmt.Errorf("marketplace: update note: %w", err)
	}
	return nil
}

func (s *MarketplaceStore) UpdateOrderDelivery(ctx context.Context, id, deliveryURL, note string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE task_orders SET delivery_url=$1, delivery_note=$2, status='delivered', updated_at=NOW() WHERE id=$3`,
		deliveryURL, note, id)
	if err != nil {
		return fmt.Errorf("marketplace: update delivery: %w", err)
	}
	return nil
}

func (s *MarketplaceStore) UpdateTaskStatus(ctx context.Context, id, status string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE marketplace_tasks SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
	if err != nil {
		return fmt.Errorf("marketplace: update task status: %w", err)
	}
	return nil
}

func (s *MarketplaceStore) ListOrdersByCreator(ctx context.Context, creatorID, status string, limit, offset int) ([]*TaskOrder, error) {
	query := `SELECT id, task_id, creator_id, advertiser_id, status, delivery_url, delivery_note, review_note, settlement_amount, created_at, updated_at
		FROM task_orders WHERE creator_id=$1`
	args := []interface{}{creatorID}
	argIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND status=$%d", argIdx)
		args = append(args, status)
		argIdx++
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
		return nil, fmt.Errorf("marketplace: list creator orders: %w", err)
	}
	defer rows.Close()

	var orders []*TaskOrder
	for rows.Next() {
		o := &TaskOrder{}
		if err := rows.Scan(&o.ID, &o.TaskID, &o.CreatorID, &o.AdvertiserID, &o.Status,
			&o.DeliveryURL, &o.DeliveryNote, &o.ReviewNote, &o.SettlementAmount,
			&o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("marketplace: scan order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (s *MarketplaceStore) ListOrdersByAdvertiser(ctx context.Context, advertiserID, status string, limit, offset int) ([]*TaskOrder, error) {
	query := `SELECT id, task_id, creator_id, advertiser_id, status, delivery_url, delivery_note, review_note, settlement_amount, created_at, updated_at
		FROM task_orders WHERE advertiser_id=$1`
	args := []interface{}{advertiserID}
	argIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND status=$%d", argIdx)
		args = append(args, status)
		argIdx++
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
		return nil, fmt.Errorf("marketplace: list advertiser orders: %w", err)
	}
	defer rows.Close()

	var orders []*TaskOrder
	for rows.Next() {
		o := &TaskOrder{}
		if err := rows.Scan(&o.ID, &o.TaskID, &o.CreatorID, &o.AdvertiserID, &o.Status,
			&o.DeliveryURL, &o.DeliveryNote, &o.ReviewNote, &o.SettlementAmount,
			&o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("marketplace: scan order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, nil
}
