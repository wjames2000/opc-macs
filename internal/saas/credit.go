package saas

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"
)

type CreditScore struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	TenantID          string    `json:"tenant_id"`
	Score             int       `json:"score"`
	OrderCompleteRate float64   `json:"order_complete_rate"`
	AvgDeliveryDays   float64   `json:"avg_delivery_days"`
	TotalOrders       int       `json:"total_orders"`
	CompletedOrders   int       `json:"completed_orders"`
	RatingAvg         float64   `json:"rating_avg"`
	RatingCount       int       `json:"rating_count"`
	Level             string    `json:"level"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type TaskReview struct {
	ID         string           `json:"id"`
	TaskID     string           `json:"task_id"`
	OrderID    string           `json:"order_id"`
	FromUserID string           `json:"from_user_id"`
	ToUserID   string           `json:"to_user_id"`
	TenantID   string           `json:"tenant_id"`
	Rating     int              `json:"rating"`
	Content    string           `json:"content"`
	Dimensions ReviewDimensions `json:"dimensions"`
	CreatedAt  time.Time        `json:"created_at"`
}

type ReviewDimensions struct {
	Quality       int `json:"quality"`
	Speed         int `json:"speed"`
	Communication int `json:"communication"`
}

type CreditStore struct {
	db *sql.DB
}

func NewCreditStore(db *sql.DB) *CreditStore {
	return &CreditStore{db: db}
}

func (s *CreditStore) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS credit_scores (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL REFERENCES users(id),
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		score INT NOT NULL DEFAULT 1000,
		order_complete_rate DECIMAL(5,4) NOT NULL DEFAULT 0,
		avg_delivery_days DECIMAL(8,2) NOT NULL DEFAULT 0,
		total_orders INT NOT NULL DEFAULT 0,
		completed_orders INT NOT NULL DEFAULT 0,
		rating_avg DECIMAL(3,2) NOT NULL DEFAULT 0,
		rating_count INT NOT NULL DEFAULT 0,
		level TEXT NOT NULL DEFAULT 'bronze',
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(tenant_id, user_id)
	);
	CREATE TABLE IF NOT EXISTS task_reviews (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL REFERENCES marketplace_tasks(id),
		order_id TEXT NOT NULL REFERENCES task_orders(id),
		from_user_id TEXT NOT NULL REFERENCES users(id),
		to_user_id TEXT NOT NULL REFERENCES users(id),
		tenant_id TEXT NOT NULL REFERENCES tenants(id),
		rating INT NOT NULL CHECK(rating>=1 AND rating<=5),
		content TEXT NOT NULL DEFAULT '',
		quality INT NOT NULL DEFAULT 5,
		speed INT NOT NULL DEFAULT 5,
		communication INT NOT NULL DEFAULT 5,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_reviews_to_user ON task_reviews(to_user_id);
	CREATE INDEX IF NOT EXISTS idx_reviews_task ON task_reviews(task_id);
	`
	_, err := s.db.Exec(schema)
	return err
}

type LevelRange struct {
	MinScore int
	Level    string
}

var creditLevels = []LevelRange{
	{MinScore: 2000, Level: "diamond"},
	{MinScore: 1500, Level: "platinum"},
	{MinScore: 1000, Level: "gold"},
	{MinScore: 600, Level: "silver"},
	{MinScore: 0, Level: "bronze"},
}

func computeLevel(score int) string {
	for _, l := range creditLevels {
		if score >= l.MinScore {
			return l.Level
		}
	}
	return "bronze"
}

func (s *CreditStore) AddReview(ctx context.Context, r *TaskReview) error {
	r.CreatedAt = time.Now()
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO task_reviews (id, task_id, order_id, from_user_id, to_user_id, tenant_id, rating, content, quality, speed, communication, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		r.ID, r.TaskID, r.OrderID, r.FromUserID, r.ToUserID, r.TenantID, r.Rating, r.Content, r.Dimensions.Quality, r.Dimensions.Speed, r.Dimensions.Communication, r.CreatedAt)
	if err != nil {
		return fmt.Errorf("credit: add review: %w", err)
	}
	return s.recalculateScore(ctx, r.ToUserID, r.TenantID)
}

func (s *CreditStore) recalculateScore(ctx context.Context, userID, tenantID string) error {
	var stats struct {
		AvgRating   float64
		RatingCount int
		Total       int
		Completed   int
		AvgDays     float64
	}

	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(AVG(rating), 0), COUNT(*),
				COALESCE(SUM(rating), 0)
		FROM task_reviews WHERE to_user_id=$1 AND tenant_id=$2`,
		userID, tenantID).Scan(&stats.AvgRating, &stats.RatingCount, &stats.Total)
	if err != nil {
		return fmt.Errorf("credit: query reviews: %w", err)
	}

	err = s.db.QueryRowContext(ctx,
		`SELECT COUNT(*),
			COALESCE(SUM(CASE WHEN status='completed' THEN 1 ELSE 0 END), 0),
			COALESCE(AVG(
				CASE WHEN completed_at IS NOT NULL AND created_at IS NOT NULL
				THEN EXTRACT(EPOCH FROM (completed_at - created_at))/86400.0
				ELSE NULL END
			), 0)
		FROM task_orders WHERE creator_id=$1 AND tenant_id=$2`,
		userID, tenantID).Scan(&stats.Total, &stats.Completed, &stats.AvgDays)
	if err != nil {
		return fmt.Errorf("credit: query orders: %w", err)
	}

	completeRate := 0.0
	if stats.Total > 0 {
		completeRate = float64(stats.Completed) / float64(stats.Total)
	}

	score := computeScore(stats.AvgRating, completeRate, stats.AvgDays, stats.Completed)
	level := computeLevel(score)

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO credit_scores (id, user_id, tenant_id, score, order_complete_rate, avg_delivery_days, total_orders, completed_orders, rating_avg, rating_count, level, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET
			score=$4, order_complete_rate=$5, avg_delivery_days=$6,
			total_orders=$7, completed_orders=$8, rating_avg=$9,
			rating_count=$10, level=$11, updated_at=$12`,
		GenerateID(), userID, tenantID, score, completeRate, stats.AvgDays, stats.Total, stats.Completed, stats.AvgRating, stats.RatingCount, level, time.Now())
	return err
}

func computeScore(avgRating, completeRate, avgDays float64, completedOrders int) int {
	ratingScore := int(avgRating * 100)
	completeScore := int(completeRate * 500)
	promptnessScore := 200
	if avgDays > 0 {
		promptnessScore = int(math.Max(0, 200-avgDays*20))
	}
	volumeBonus := int(math.Min(float64(completedOrders)*10, 200))
	return ratingScore + completeScore + promptnessScore + volumeBonus
}

func (s *CreditStore) GetScore(ctx context.Context, tenantID, userID string) (*CreditScore, error) {
	cs := &CreditScore{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, tenant_id, score, order_complete_rate, avg_delivery_days,
			total_orders, completed_orders, rating_avg, rating_count, level, updated_at
		FROM credit_scores WHERE tenant_id=$1 AND user_id=$2`,
		tenantID, userID).Scan(&cs.ID, &cs.UserID, &cs.TenantID, &cs.Score, &cs.OrderCompleteRate, &cs.AvgDeliveryDays,
		&cs.TotalOrders, &cs.CompletedOrders, &cs.RatingAvg, &cs.RatingCount, &cs.Level, &cs.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("credit: get score: %w", err)
	}
	return cs, nil
}
