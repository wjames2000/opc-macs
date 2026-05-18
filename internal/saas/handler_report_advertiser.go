package saas

import (
	"context"
	"database/sql"
	"net/http"
)

type AdvertiserReportStore struct {
	db *sql.DB
}

func NewAdvertiserReportStore(db *sql.DB) *AdvertiserReportStore {
	return &AdvertiserReportStore{db: db}
}

func (s *AdvertiserReportStore) GetSpendingTotal(ctx context.Context, advertiserID, startDate, endDate string) (float64, error) {
	var total sql.NullFloat64
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(budget), 0) FROM tasks
		 WHERE advertiser_id = $1 AND created_at >= $2 AND created_at <= $3`,
		advertiserID, startDate, endDate).Scan(&total)
	return total.Float64, err
}

func (s *AdvertiserReportStore) GetTaskCount(ctx context.Context, advertiserID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tasks WHERE advertiser_id = $1`, advertiserID).Scan(&count)
	return count, err
}

func (s *AdvertiserReportStore) GetOrderStatusBreakdown(ctx context.Context, advertiserID string) ([]StatusCount, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT o.status, COUNT(*) FROM orders o
		 JOIN tasks t ON o.task_id = t.id
		 WHERE t.advertiser_id = $1 GROUP BY o.status ORDER BY COUNT(*) DESC`,
		advertiserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var breakdown []StatusCount
	for rows.Next() {
		var sc StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, err
		}
		breakdown = append(breakdown, sc)
	}
	return breakdown, rows.Err()
}

func (s *AdvertiserReportStore) GetTopTasks(ctx context.Context, advertiserID string, limit int) ([]TopTask, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT t.id, t.title, t.budget, t.platform,
		 COALESCE(oc.order_count, 0) as order_count
		 FROM tasks t LEFT JOIN (
		   SELECT task_id, COUNT(*) as order_count FROM orders GROUP BY task_id
		 ) oc ON t.id = oc.task_id
		 WHERE t.advertiser_id = $1 ORDER BY t.budget DESC LIMIT $2`,
		advertiserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []TopTask
	for rows.Next() {
		var tt TopTask
		if err := rows.Scan(&tt.ID, &tt.Title, &tt.Budget, &tt.Platform, &tt.OrderCount); err != nil {
			return nil, err
		}
		tasks = append(tasks, tt)
	}
	return tasks, rows.Err()
}

type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type TopTask struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Budget     float64 `json:"budget"`
	Platform   string  `json:"platform"`
	OrderCount int     `json:"order_count"`
}

type AdvertiserReportHandler struct {
	store *AdvertiserReportStore
}

func NewAdvertiserReportHandler(store *AdvertiserReportStore) *AdvertiserReportHandler {
	return &AdvertiserReportHandler{store: store}
}

func (h *AdvertiserReportHandler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	_, userID := getTenantID(r), getUserID(r)
	startDate := r.URL.Query().Get("start_date")
	if startDate == "" {
		startDate = "2000-01-01"
	}
	endDate := r.URL.Query().Get("end_date")
	if endDate == "" {
		endDate = "2099-12-31"
	}

	totalSpent, err := h.store.GetSpendingTotal(r.Context(), userID, startDate, endDate)
	if err != nil {
		jsonError(w, 500, "spending total: "+err.Error())
		return
	}

	totalTasks, err := h.store.GetTaskCount(r.Context(), userID)
	if err != nil {
		jsonError(w, 500, "task count: "+err.Error())
		return
	}

	breakdown, err := h.store.GetOrderStatusBreakdown(r.Context(), userID)
	if err != nil {
		jsonError(w, 500, "order breakdown: "+err.Error())
		return
	}
	if breakdown == nil {
		breakdown = []StatusCount{}
	}

	topTasks, err := h.store.GetTopTasks(r.Context(), userID, 10)
	if err != nil {
		jsonError(w, 500, "top tasks: "+err.Error())
		return
	}
	if topTasks == nil {
		topTasks = []TopTask{}
	}

	totalOrders := 0
	for _, b := range breakdown {
		totalOrders += b.Count
	}

	jsonResponse(w, 200, map[string]interface{}{
		"advertiser_id":   userID,
		"total_spent":     totalSpent,
		"total_tasks":     totalTasks,
		"total_orders":    totalOrders,
		"period_start":    startDate,
		"period_end":      endDate,
		"order_breakdown": breakdown,
		"top_tasks":       topTasks,
	})
}
