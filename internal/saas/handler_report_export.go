package saas

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type ReportExportStore struct {
	db *sql.DB
}

func NewReportExportStore(db *sql.DB) *ReportExportStore {
	return &ReportExportStore{db: db}
}

func (s *ReportExportStore) ExportOrdersCSV(ctx context.Context, advertiserID string, startDate, endDate string) ([][]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT o.id, o.task_id, t.title, o.creator_id, o.status, o.bid_amount, o.created_at
		 FROM orders o JOIN tasks t ON o.task_id = t.id
		 WHERE t.advertiser_id = $1 AND o.created_at >= $2 AND o.created_at <= $3
		 ORDER BY o.created_at DESC`,
		advertiserID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := [][]string{
		{"订单ID", "任务ID", "任务标题", "创作者ID", "状态", "金额", "创建时间"},
	}
	for rows.Next() {
		var id, taskID, title, creatorID, status string
		var bidAmount float64
		var createdAt time.Time
		if err := rows.Scan(&id, &taskID, &title, &creatorID, &status, &bidAmount, &createdAt); err != nil {
			return nil, err
		}
		records = append(records, []string{
			id, taskID, title, creatorID, status,
			strconv.FormatFloat(bidAmount, 'f', 2, 64),
			createdAt.Format("2006-01-02 15:04:05"),
		})
	}
	return records, rows.Err()
}

func (s *ReportExportStore) ExportTasksCSV(ctx context.Context, advertiserID string) ([][]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, platform, budget, status, created_at
		 FROM tasks WHERE advertiser_id = $1 ORDER BY created_at DESC`,
		advertiserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := [][]string{
		{"任务ID", "标题", "平台", "预算", "状态", "创建时间"},
	}
	for rows.Next() {
		var id, title, platform, status string
		var budget float64
		var createdAt time.Time
		if err := rows.Scan(&id, &title, &platform, &budget, &status, &createdAt); err != nil {
			return nil, err
		}
		records = append(records, []string{
			id, title, platform,
			strconv.FormatFloat(budget, 'f', 2, 64),
			status,
			createdAt.Format("2006-01-02 15:04:05"),
		})
	}
	return records, rows.Err()
}

type ReportExportHandler struct {
	store *ReportExportStore
}

func NewReportExportHandler(store *ReportExportStore) *ReportExportHandler {
	return &ReportExportHandler{store: store}
}

func (h *ReportExportHandler) HandleExportOrders(w http.ResponseWriter, r *http.Request) {
	_, userID := getTenantID(r), getUserID(r)
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if startDate == "" {
		startDate = "2000-01-01"
	}
	if endDate == "" {
		endDate = "2099-12-31"
	}

	records, err := h.store.ExportOrdersCSV(r.Context(), userID, startDate, endDate)
	if err != nil {
		jsonError(w, 500, "export orders: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=orders_%s.csv", time.Now().Format("20060102")))
	writer := csv.NewWriter(w)
	for _, rec := range records {
		writer.Write(rec)
	}
	writer.Flush()
}

func (h *ReportExportHandler) HandleExportTasks(w http.ResponseWriter, r *http.Request) {
	_, userID := getTenantID(r), getUserID(r)

	records, err := h.store.ExportTasksCSV(r.Context(), userID)
	if err != nil {
		jsonError(w, 500, "export tasks: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=tasks_%s.csv", time.Now().Format("20060102")))
	writer := csv.NewWriter(w)
	for _, rec := range records {
		writer.Write(rec)
	}
	writer.Flush()
}
