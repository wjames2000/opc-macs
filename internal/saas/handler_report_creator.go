package saas

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

type CreatorReportStore struct {
	db *sql.DB
}

func NewCreatorReportStore(db *sql.DB) *CreatorReportStore {
	return &CreatorReportStore{db: db}
}

func (s *CreatorReportStore) GetEarningsTotal(ctx context.Context, creatorID string) (float64, float64, error) {
	var earned, pending sql.NullFloat64
	err := s.db.QueryRowContext(ctx,
		`SELECT
		 COALESCE(SUM(CASE WHEN status = 'completed' THEN bid_amount ELSE 0 END), 0),
		 COALESCE(SUM(CASE WHEN status IN ('pending','approved') THEN bid_amount ELSE 0 END), 0)
		 FROM orders WHERE creator_id = $1`, creatorID).Scan(&earned, &pending)
	return earned.Float64, pending.Float64, err
}

func (s *CreatorReportStore) GetCompletedCount(ctx context.Context, creatorID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM orders WHERE creator_id = $1 AND status = 'completed'`,
		creatorID).Scan(&count)
	return count, err
}

func (s *CreatorReportStore) GetCompletedOrders(ctx context.Context, creatorID string, limit, offset int) ([]CompletedOrder, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT o.id, o.task_id, t.title, t.budget, t.platform, o.status, o.updated_at
		 FROM orders o JOIN tasks t ON o.task_id = t.id
		 WHERE o.creator_id = $1 AND o.status = 'completed'
		 ORDER BY o.updated_at DESC LIMIT $2 OFFSET $3`,
		creatorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []CompletedOrder
	for rows.Next() {
		var co CompletedOrder
		if err := rows.Scan(&co.ID, &co.TaskID, &co.TaskTitle, &co.Budget,
			&co.Platform, &co.Status, &co.CompletedAt); err != nil {
			return nil, err
		}
		orders = append(orders, co)
	}
	return orders, rows.Err()
}

func (s *CreatorReportStore) GetPlatformBreakdown(ctx context.Context, creatorID string) ([]PlatformEarning, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT t.platform, COUNT(*), COALESCE(SUM(o.bid_amount), 0)
		 FROM orders o JOIN tasks t ON o.task_id = t.id
		 WHERE o.creator_id = $1 AND o.status = 'completed'
		 GROUP BY t.platform ORDER BY SUM(o.bid_amount) DESC`,
		creatorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var breakdown []PlatformEarning
	for rows.Next() {
		var pe PlatformEarning
		if err := rows.Scan(&pe.Platform, &pe.Count, &pe.Total); err != nil {
			return nil, err
		}
		breakdown = append(breakdown, pe)
	}
	return breakdown, rows.Err()
}

func (s *CreatorReportStore) GetMonthlyEarnings(ctx context.Context, creatorID string, months int) ([]MonthlyEarning, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT TO_CHAR(o.updated_at, 'YYYY-MM') as month,
		 COUNT(*), COALESCE(SUM(o.bid_amount), 0)
		 FROM orders o
		 WHERE o.creator_id = $1 AND o.status = 'completed'
		   AND o.updated_at > $2
		 GROUP BY month ORDER BY month DESC LIMIT $3`,
		creatorID, time.Now().AddDate(0, -months, 0), months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var earnings []MonthlyEarning
	for rows.Next() {
		var me MonthlyEarning
		if err := rows.Scan(&me.Month, &me.Count, &me.Total); err != nil {
			return nil, err
		}
		earnings = append(earnings, me)
	}
	return earnings, rows.Err()
}

type CompletedOrder struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	TaskTitle   string    `json:"task_title"`
	Budget      float64   `json:"budget"`
	Platform    string    `json:"platform"`
	Status      string    `json:"status"`
	CompletedAt time.Time `json:"completed_at"`
}

type PlatformEarning struct {
	Platform string  `json:"platform"`
	Count    int     `json:"count"`
	Total    float64 `json:"total"`
}

type MonthlyEarning struct {
	Month string  `json:"month"`
	Count int     `json:"count"`
	Total float64 `json:"total"`
}

type CreatorReportHandler struct {
	store *CreatorReportStore
}

func NewCreatorReportHandler(store *CreatorReportStore) *CreatorReportHandler {
	return &CreatorReportHandler{store: store}
}

func (h *CreatorReportHandler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	_, userID := getTenantID(r), getUserID(r)
	months := parseInt(r.URL.Query().Get("months"), 6)
	if months > 24 {
		months = 24
	}

	earned, pending, err := h.store.GetEarningsTotal(r.Context(), userID)
	if err != nil {
		jsonError(w, 500, "earnings total: "+err.Error())
		return
	}

	completedCount, err := h.store.GetCompletedCount(r.Context(), userID)
	if err != nil {
		jsonError(w, 500, "completed count: "+err.Error())
		return
	}

	breakdown, err := h.store.GetPlatformBreakdown(r.Context(), userID)
	if err != nil {
		jsonError(w, 500, "platform breakdown: "+err.Error())
		return
	}
	if breakdown == nil {
		breakdown = []PlatformEarning{}
	}

	monthly, err := h.store.GetMonthlyEarnings(r.Context(), userID, months)
	if err != nil {
		jsonError(w, 500, "monthly earnings: "+err.Error())
		return
	}
	if monthly == nil {
		monthly = []MonthlyEarning{}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"creator_id":         userID,
		"total_earned":       earned,
		"pending_amount":     pending,
		"completed_orders":   completedCount,
		"platform_breakdown": breakdown,
		"monthly_earnings":   monthly,
	})
}

func (h *CreatorReportHandler) HandleOrders(w http.ResponseWriter, r *http.Request) {
	_, userID := getTenantID(r), getUserID(r)
	limit := parseInt(r.URL.Query().Get("limit"), 10)
	if limit > 100 {
		limit = 100
	}
	offset := parseInt(r.URL.Query().Get("offset"), 0)

	orders, err := h.store.GetCompletedOrders(r.Context(), userID, limit, offset)
	if err != nil {
		jsonError(w, 500, "list orders: "+err.Error())
		return
	}
	if orders == nil {
		orders = []CompletedOrder{}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"orders": orders,
		"total":  len(orders),
	})
}

func (h *CreatorReportHandler) HandleMonthly(w http.ResponseWriter, r *http.Request) {
	_, userID := getTenantID(r), getUserID(r)
	months := parseInt(r.URL.Query().Get("months"), 12)
	if months > 24 {
		months = 24
	}

	monthly, err := h.store.GetMonthlyEarnings(r.Context(), userID, months)
	if err != nil {
		jsonError(w, 500, "monthly earnings: "+err.Error())
		return
	}
	if monthly == nil {
		monthly = []MonthlyEarning{}
	}

	jsonResponse(w, 200, map[string]interface{}{
		"monthly_earnings": monthly,
		"total":            len(monthly),
	})
}
