package saas

import (
	"encoding/json"
	"net/http"
)

type CreditHandler struct {
	credit *CreditStore
}

func NewCreditHandler(credit *CreditStore) *CreditHandler {
	return &CreditHandler{credit: credit}
}

type addReviewRequest struct {
	TaskID        string `json:"task_id"`
	OrderID       string `json:"order_id"`
	ToUserID      string `json:"to_user_id"`
	Rating        int    `json:"rating"`
	Content       string `json:"content"`
	Quality       int    `json:"quality"`
	Speed         int    `json:"speed"`
	Communication int    `json:"communication"`
}

func (h *CreditHandler) HandleAddReview(w http.ResponseWriter, r *http.Request) {
	tenantID, fromUserID := r.Context().Value(CtxTenantID).(string), r.Context().Value(CtxUserID).(string)

	var req addReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.TaskID == "" || req.OrderID == "" || req.ToUserID == "" {
		writeError(w, http.StatusBadRequest, "task_id, order_id, to_user_id required")
		return
	}
	if req.Rating < 1 || req.Rating > 5 {
		writeError(w, http.StatusBadRequest, "rating must be 1-5")
		return
	}

	review := &TaskReview{
		ID:         GenerateID(),
		TaskID:     req.TaskID,
		OrderID:    req.OrderID,
		FromUserID: fromUserID,
		ToUserID:   req.ToUserID,
		TenantID:   tenantID,
		Rating:     req.Rating,
		Content:    req.Content,
		Dimensions: ReviewDimensions{
			Quality:       clampInt(req.Quality, 1, 5),
			Speed:         clampInt(req.Speed, 1, 5),
			Communication: clampInt(req.Communication, 1, 5),
		},
	}

	if err := h.credit.AddReview(r.Context(), review); err != nil {
		writeError(w, http.StatusInternalServerError, "add review: "+err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, review)
}

func (h *CreditHandler) HandleGetScore(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Context().Value(CtxTenantID).(string)
	targetUser := r.PathValue("user_id")
	if targetUser == "" {
		writeError(w, http.StatusBadRequest, "user_id required")
		return
	}

	score, err := h.credit.GetScore(r.Context(), tenantID, targetUser)
	if err != nil {
		writeError(w, http.StatusNotFound, "score not found")
		return
	}
	writeJSON(w, http.StatusOK, score)
}

func (h *CreditHandler) HandleGetMyScore(w http.ResponseWriter, r *http.Request) {
	tenantID, userID := r.Context().Value(CtxTenantID).(string), r.Context().Value(CtxUserID).(string)
	score, err := h.credit.GetScore(r.Context(), tenantID, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "score not found")
		return
	}
	writeJSON(w, http.StatusOK, score)
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
