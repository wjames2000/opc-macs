package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type listToolsResponse struct {
	Tools []ToolDefinition `json:"tools"`
}

type executeRequest struct {
	Name   string          `json:"name"`
	Params json.RawMessage `json:"params,omitempty"`
}

type executeResponse struct {
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

// handleListTools handles GET /mcp/v1/tools - returns all registered tools.
func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := s.ListTools()
	resp := listToolsResponse{Tools: tools}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleExecuteTool handles POST /mcp/v1/tools/{name} - executes a specific tool.
func (s *Server) handleExecuteTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tool name from URL path: /mcp/v1/tools/{name}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/mcp/v1/tools/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "tool name required", http.StatusBadRequest)
		return
	}
	toolName := parts[0]

	var req executeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		req.Name = toolName
	}

	result, err := s.ExecuteTool(r.Context(), req.Name, req.Params)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(executeResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(executeResponse{Result: result})
}

// handleSSE handles GET /mcp/v1/sse - SSE connection for streaming responses.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Send initial tools list event
	tools := s.ListTools()
	toolsJSON, _ := json.Marshal(tools)
	_, _ = fmt.Fprintf(w, "event: tools\ndata: %s\n\n", toolsJSON)
	flusher.Flush()

	// Keep connection alive
	<-r.Context().Done()
}
