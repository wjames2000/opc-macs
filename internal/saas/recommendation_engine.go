package saas

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
)

type RecommendationHandler struct {
	store *MarketplaceStore
}

func NewRecommendationHandler(store *MarketplaceStore) *RecommendationHandler {
	return &RecommendationHandler{store: store}
}

type TaskRecommendation struct {
	Task   *MarketplaceTask `json:"task"`
	Score  float64          `json:"score"`
	Reason string           `json:"reason"`
}

func (h *RecommendationHandler) HandleRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		jsonError(w, 405, "method not allowed")
		return
	}

	userID := getUserID(r)
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	tasks, err := h.store.ListOpenTasks(r.Context(), "", "", 100, 0)
	if err != nil {
		jsonError(w, 500, "list tasks: "+err.Error())
		return
	}

	userOrders, _ := h.store.ListOrdersByCreator(r.Context(), userID, "", 100, 0)
	appliedTaskIDs := make(map[string]bool)
	for _, o := range userOrders {
		appliedTaskIDs[o.TaskID] = true
	}

	var recs []TaskRecommendation
	for _, t := range tasks {
		if appliedTaskIDs[t.ID] {
			continue
		}
		score := t.Budget / 100.0
		if t.Budget > 10000 {
			score *= 1.5
		}
		if t.Budget > 50000 {
			score *= 2.0
		}
		score = math.Round(score*100) / 100

		reason := fmt.Sprintf("预算 ¥%.0f", t.Budget)
		if t.Budget > 50000 {
			reason += " · 高价值任务"
		}
		if t.Platform != "" {
			reason += fmt.Sprintf(" · %s", t.Platform)
		}

		recs = append(recs, TaskRecommendation{
			Task:   t,
			Score:  score,
			Reason: reason,
		})
	}

	for i := 0; i < len(recs); i++ {
		for j := i + 1; j < len(recs); j++ {
			if recs[j].Score > recs[i].Score {
				recs[i], recs[j] = recs[j], recs[i]
			}
		}
	}

	if len(recs) > limit {
		recs = recs[:limit]
	}

	jsonResponse(w, 200, map[string]interface{}{
		"recommendations": recs,
		"total":           len(recs),
	})
}

func (h *RecommendationHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, 200, map[string]interface{}{
		"name":        "recommendation_engine",
		"version":     "1.0.0",
		"description": "创作者任务推荐引擎，基于预算、平台匹配度推荐最优任务",
		"algorithm":   "budget_weighted + platform_match",
	})
}
