package saas

import (
	"encoding/json"
	"net/http"
)

type ReviewHandler struct {
	store *MarketplaceStore
}

func NewReviewHandler(store *MarketplaceStore) *ReviewHandler {
	return &ReviewHandler{store: store}
}

func (h *ReviewHandler) HandleApprove(w http.ResponseWriter, r *http.Request) {
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
		Note string `json:"note"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.store.UpdateOrderStatus(r.Context(), orderID, "approved"); err != nil {
		jsonError(w, 500, "approve failed: "+err.Error())
		return
	}
	if req.Note != "" {
		h.store.UpdateOrderNote(r.Context(), orderID, req.Note)
	}

	jsonResponse(w, 200, map[string]string{"status": "approved"})
}

func (h *ReviewHandler) HandleReject(w http.ResponseWriter, r *http.Request) {
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
		Note string `json:"note"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := h.store.UpdateOrderStatus(r.Context(), orderID, "rejected"); err != nil {
		jsonError(w, 500, "reject failed: "+err.Error())
		return
	}
	if req.Note != "" {
		h.store.UpdateOrderNote(r.Context(), orderID, req.Note)
	}

	jsonResponse(w, 200, map[string]string{"status": "rejected"})
}
