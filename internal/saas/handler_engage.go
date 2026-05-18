package saas

import (
	"encoding/json"
	"net/http"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type EngageHandler struct {
	loader *runtime.Loader
}

func NewEngageHandler(loader *runtime.Loader) *EngageHandler {
	return &EngageHandler{loader: loader}
}

func (h *EngageHandler) HandleReply(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}

	var req struct {
		Platform  string `json:"platform"`
		PostID    string `json:"post_id"`
		CommentID string `json:"comment_id"`
		Content   string `json:"content"`
		Template  string `json:"template"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}
	if req.Platform == "" || req.Content == "" {
		jsonError(w, 400, "platform and content are required")
		return
	}

	plugin, ok := h.loader.Get("engage")
	if !ok {
		jsonError(w, 404, "engage agent not found")
		return
	}

	input, _ := json.Marshal(map[string]interface{}{
		"action":     "reply",
		"platform":   req.Platform,
		"post_id":    req.PostID,
		"comment_id": req.CommentID,
		"content":    req.Content,
		"template":   req.Template,
	})
	result, err := plugin.Execute(r.Context(), string(input), map[string]interface{}{})
	if err != nil {
		jsonError(w, 500, "reply failed: "+err.Error())
		return
	}

	jsonResponse(w, 200, result.Data)
}

func (h *EngageHandler) HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}

	var req struct {
		Platform string `json:"platform"`
		PostID   string `json:"post_id"`
		Period   string `json:"period"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}
	if req.Platform == "" {
		jsonError(w, 400, "platform is required")
		return
	}

	plugin, ok := h.loader.Get("engage")
	if !ok {
		jsonError(w, 404, "engage agent not found")
		return
	}

	input, _ := json.Marshal(map[string]interface{}{
		"action":   "analyze",
		"platform": req.Platform,
		"post_id":  req.PostID,
		"period":   req.Period,
	})
	result, err := plugin.Execute(r.Context(), string(input), map[string]interface{}{})
	if err != nil {
		jsonError(w, 500, "analysis failed: "+err.Error())
		return
	}

	jsonResponse(w, 200, result.Data)
}

func (h *EngageHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	agents := h.loader.List()
	for _, a := range agents {
		if a.Name == "engage" {
			jsonResponse(w, 200, a)
			return
		}
	}
	jsonError(w, 404, "plugin not found")
}
