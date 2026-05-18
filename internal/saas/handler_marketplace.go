package saas

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// MarketplaceHandlers provides HTTP handlers for marketplace operations.
type MarketplaceHandlers struct {
	store *MarketplaceStore
}

func NewMarketplaceHandlers(store *MarketplaceStore) *MarketplaceHandlers {
	return &MarketplaceHandlers{store: store}
}

// POST /api/v1/marketplace/tasks
func (h *MarketplaceHandlers) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}

	var req struct {
		Title          string  `json:"title"`
		Description    string  `json:"description"`
		Platform       string  `json:"platform"`
		ContentType    string  `json:"content_type"`
		SettlementType string  `json:"settlement_type"`
		Budget         float64 `json:"budget"`
		Requirements   string  `json:"requirements"`
		Deadline       string  `json:"deadline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}
	if req.Title == "" {
		jsonError(w, 400, "title is required")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	userID := r.Header.Get("X-User-ID")

	deadline, _ := time.Parse(time.RFC3339, req.Deadline)

	task := &MarketplaceTask{
		ID:             GenerateID(),
		TenantID:       tenantID,
		CreatorID:      userID,
		Title:          req.Title,
		Description:    req.Description,
		Platform:       req.Platform,
		ContentType:    req.ContentType,
		SettlementType: req.SettlementType,
		Budget:         req.Budget,
		Requirements:   req.Requirements,
		Status:         "open",
		Deadline:       deadline,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.store.CreateTask(r.Context(), task); err != nil {
		jsonError(w, 500, "create task failed: "+err.Error())
		return
	}

	jsonResponse(w, 201, task)
}

// GET /api/v1/marketplace/tasks
func (h *MarketplaceHandlers) HandleListMyTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonError(w, 405, "method not allowed")
		return
	}

	tenantID := r.Header.Get("X-Tenant-ID")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	tasks, err := h.store.ListOpenTasks(r.Context(), "", "", limit, offset)
	if err != nil {
		jsonError(w, 500, "list tasks failed: "+err.Error())
		return
	}

	_ = tenantID
	jsonResponse(w, 200, tasks)
}

// GET /api/v1/marketplace/tasks/{id}
func (h *MarketplaceHandlers) HandleGetTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonError(w, 405, "method not allowed")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "task id required")
		return
	}

	task, err := h.store.GetTask(r.Context(), id)
	if err != nil {
		jsonError(w, 404, "task not found")
		return
	}

	jsonResponse(w, 200, task)
}

// GET /api/v1/marketplace/open-tasks
func (h *MarketplaceHandlers) HandleListOpenTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonError(w, 405, "method not allowed")
		return
	}

	platform := r.URL.Query().Get("platform")
	contentType := r.URL.Query().Get("content_type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	tasks, err := h.store.ListOpenTasks(r.Context(), platform, contentType, limit, offset)
	if err != nil {
		jsonError(w, 500, "list open tasks failed: "+err.Error())
		return
	}

	jsonResponse(w, 200, tasks)
}

// POST /api/v1/marketplace/tasks/{id}/apply
func (h *MarketplaceHandlers) HandleApplyTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}

	taskID := r.PathValue("id")
	userID := r.Header.Get("X-User-ID")
	tenantID := r.Header.Get("X-Tenant-ID")

	task, err := h.store.GetTask(r.Context(), taskID)
	if err != nil {
		jsonError(w, 404, "task not found")
		return
	}

	order := &TaskOrder{
		ID:           GenerateID(),
		TaskID:       taskID,
		CreatorID:    userID,
		AdvertiserID: task.CreatorID,
		Status:       "applied",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.store.CreateOrder(r.Context(), order); err != nil {
		jsonError(w, 409, "apply failed: "+err.Error())
		return
	}

	_ = tenantID
	jsonResponse(w, 201, order)
}

// GET /api/v1/marketplace/my-orders
func (h *MarketplaceHandlers) HandleListMyOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonError(w, 405, "method not allowed")
		return
	}

	userID := r.Header.Get("X-User-ID")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	orders, err := h.store.ListOrdersByCreator(r.Context(), userID, status, limit, offset)
	if err != nil {
		jsonError(w, 500, "list orders failed: "+err.Error())
		return
	}

	jsonResponse(w, 200, orders)
}

func (h *MarketplaceHandlers) HandleCancelTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		jsonError(w, 400, "task id required")
		return
	}

	if err := h.store.UpdateTaskStatus(r.Context(), id, "cancelled"); err != nil {
		jsonError(w, 500, "cancel task failed: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]string{"status": "cancelled"})
}
