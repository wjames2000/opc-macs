package saas

import (
	"encoding/json"
	"net/http"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type TrendRadarHandler struct {
	loader *runtime.Loader
}

func NewTrendRadarHandler(loader *runtime.Loader) *TrendRadarHandler {
	return &TrendRadarHandler{loader: loader}
}

func (h *TrendRadarHandler) HandleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}

	var req struct {
		Platforms []string `json:"platforms"`
		Category  string   `json:"category"`
		Period    string   `json:"period"`
		Keywords  string   `json:"keywords"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}

	plugin, ok := h.loader.Get("trend_radar")
	if !ok {
		jsonError(w, 404, "trend_radar agent not found")
		return
	}

	input, _ := json.Marshal(req)
	result, err := plugin.Execute(r.Context(), string(input), map[string]interface{}{})
	if err != nil {
		jsonError(w, 500, "generate trend report failed: "+err.Error())
		return
	}

	jsonResponse(w, 200, result.Data)
}

func (h *TrendRadarHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	agents := h.loader.List()
	for _, a := range agents {
		if a.Name == "trend_radar" {
			jsonResponse(w, 200, a)
			return
		}
	}
	jsonError(w, 404, "plugin not found")
}
