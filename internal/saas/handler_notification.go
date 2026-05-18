package saas

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type NotificationHandler struct {
	store     *NotificationStore
	lifecycle *OrderLifecycle
}

func NewNotificationHandler(store *NotificationStore, lifecycle *OrderLifecycle) *NotificationHandler {
	return &NotificationHandler{
		store:     store,
		lifecycle: lifecycle,
	}
}

func notifID() string {
	return fmt.Sprintf("notif_%d", time.Now().UnixNano())
}

func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	if userID == "" {
		jsonError(w, 401, "unauthorized")
		return
	}
	unreadOnly := r.URL.Query().Get("unread_only") == "true"
	notifs, err := h.store.ListByUser(r.Context(), userID, unreadOnly, 50, 0)
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, notifs)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	if userID == "" {
		jsonError(w, 401, "unauthorized")
		return
	}
	id := r.PathValue("id")
	if id == "all" {
		if err := h.store.MarkAllRead(r.Context(), userID); err != nil {
			jsonError(w, 500, err.Error())
			return
		}
		jsonResponse(w, 200, map[string]string{"status": "ok"})
		return
	}
	if err := h.store.MarkRead(r.Context(), id); err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "ok"})
}

func (h *NotificationHandler) CountUnread(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	if userID == "" {
		jsonError(w, 401, "unauthorized")
		return
	}
	count, err := h.store.CountUnread(r.Context(), userID)
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, map[string]int{"count": count})
}

func (h *NotificationHandler) CompleteOrder(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		jsonError(w, 401, "unauthorized")
		return
	}
	var req struct {
		OrderID string `json:"order_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if err := h.lifecycle.CompleteOrder(r.Context(), tenantID, req.OrderID); err != nil {
		jsonError(w, 500, err.Error())
		return
	}

	h.store.Create(r.Context(), &Notification{
		ID: GenerateID(), TenantID: tenantID, UserID: userID,
		Type: "order_status", Title: "任务已完成",
		Content: "订单已完成结算，收益已发放至钱包", CreatedAt: time.Now(),
	})

	jsonResponse(w, 200, map[string]string{"status": "completed"})
}

func (h *NotificationHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderID string `json:"order_id"`
		Reason  string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	if err := h.lifecycle.CancelOrder(r.Context(), req.OrderID, req.Reason); err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "cancelled"})
}

// POST /api/v1/notifications — create notification manually
func (h *NotificationHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var n Notification
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		jsonError(w, 400, "invalid body")
		return
	}
	n.ID = notifID()
	n.CreatedAt = time.Now()
	if err := h.store.Create(r.Context(), &n); err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 201, n)
}

func (h *NotificationHandler) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.Delete(r.Context(), id); err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "deleted"})
}
