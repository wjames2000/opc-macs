package saas

import (
	"encoding/json"
	"net/http"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type TagGeneratorHandler struct {
	loader *runtime.Loader
}

func NewTagGeneratorHandler(loader *runtime.Loader) *TagGeneratorHandler {
	return &TagGeneratorHandler{loader: loader}
}

func (h *TagGeneratorHandler) HandleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		jsonError(w, 405, "method not allowed")
		return
	}

	var req struct {
		Platform string `json:"platform"`
		Content  string `json:"content"`
		Style    string `json:"style"`
		Category string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}
	if req.Content == "" {
		jsonError(w, 400, "content is required")
		return
	}

	plugin, ok := h.loader.Get("tag_generator")
	if !ok {
		jsonError(w, 404, "tag_generator agent not found")
		return
	}

	input, _ := json.Marshal(req)
	result, err := plugin.Execute(r.Context(), string(input), map[string]interface{}{})
	if err != nil {
		jsonError(w, 500, "generate tags failed: "+err.Error())
		return
	}

	jsonResponse(w, 200, result.Data)
}

func (h *TagGeneratorHandler) HandleInfo(w http.ResponseWriter, r *http.Request) {
	agents := h.loader.List()
	for _, a := range agents {
		if a.Name == "tag_generator" {
			jsonResponse(w, 200, a)
			return
		}
	}
	jsonError(w, 404, "plugin not found")
}
