package saas

import (
	"encoding/json"
	"net/http"
)

type DeliveryHandler struct {
	store *MarketplaceStore
}

func NewDeliveryHandler(store *MarketplaceStore) *DeliveryHandler {
	return &DeliveryHandler{store: store}
}

func (h *DeliveryHandler) HandleDeliver(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}

	orderID := r.PathValue("id")
	if orderID == "" {
		jsonError(w, 400, "order id required")
		return
	}

	var req struct {
		DeliveryURL  string `json:"delivery_url"`
		DeliveryNote string `json:"delivery_note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}
	if req.DeliveryURL == "" {
		jsonError(w, 400, "delivery_url is required")
		return
	}

	if err := h.store.UpdateOrderDelivery(r.Context(), orderID, req.DeliveryURL, req.DeliveryNote); err != nil {
		jsonError(w, 500, "deliver failed: "+err.Error())
		return
	}

	jsonResponse(w, 200, map[string]string{"status": "delivered"})
}

func (h *DeliveryHandler) HandleGetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonError(w, 405, "method not allowed")
		return
	}

	orderID := r.PathValue("id")
	if orderID == "" {
		jsonError(w, 400, "order id required")
		return
	}

	order, err := h.store.GetOrder(r.Context(), orderID)
	if err != nil {
		jsonError(w, 404, "order not found")
		return
	}

	jsonResponse(w, 200, order)
}
