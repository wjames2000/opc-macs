package saas

import (
	"encoding/json"
	"net/http"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type VideoScriptHandler struct {
	loader *runtime.Loader
}

func NewVideoScriptHandler(loader *runtime.Loader) *VideoScriptHandler {
	return &VideoScriptHandler{loader: loader}
}

func (h *VideoScriptHandler) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}

	var req struct {
		Platform string `json:"platform"`
		Product  string `json:"product"`
		Style    string `json:"style"`
		Duration int    `json:"duration"`
		Keywords string `json:"keywords"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}
	if req.Product == "" {
		jsonError(w, 400, "product is required")
		return
	}

	plugin, ok := h.loader.Get("video_script")
	if !ok {
		jsonError(w, 404, "video_script agent not found")
		return
	}

	input, _ := json.Marshal(req)
	result, err := plugin.Execute(r.Context(), string(input), map[string]interface{}{})
	if err != nil {
		jsonError(w, 500, "generate script failed: "+err.Error())
		return
	}

	jsonResponse(w, 200, result.Data)
}

func (h *VideoScriptHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	agents := h.loader.List()
	for _, a := range agents {
		if a.Name == "video_script" {
			jsonResponse(w, 200, a)
			return
		}
	}
	jsonError(w, 404, "plugin not found")
}
