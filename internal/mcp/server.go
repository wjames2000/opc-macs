package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// ToolHandler is the function signature for an MCP tool.
type ToolHandler func(ctx context.Context, params json.RawMessage) (any, error)

// ToolDefinition describes an MCP tool for discovery.
type ToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"input_schema"`
}

// Server implements a lightweight Model Context Protocol server.
// It supports SSE transport (Server-Sent Events for streaming responses)
// and JSON-RPC style request handling for tool discovery and execution.
type Server struct {
	mu     sync.RWMutex
	tools  map[string]ToolEntry
	cfg    Config
	server *http.Server
}

// ToolEntry holds a registered tool's definition and handler.
type ToolEntry struct {
	Definition ToolDefinition
	Handler    ToolHandler
}

// Config holds MCP server configuration.
type Config struct {
	Enabled   bool
	Port      int
	Transport string // "sse" or "stdio"
}

// NewServer creates a new MCP server with the given configuration.
func NewServer(cfg Config) *Server {
	return &Server{
		tools: make(map[string]ToolEntry),
		cfg:   cfg,
	}
}

// RegisterTool registers a tool with the MCP server.
func (s *Server) RegisterTool(name, description string, inputSchema any, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[name] = ToolEntry{
		Definition: ToolDefinition{
			Name:        name,
			Description: description,
			InputSchema: inputSchema,
		},
		Handler: handler,
	}
	log.Printf("[MCP] Registered tool: %s", name)
}

// ListTools returns all registered tool definitions.
func (s *Server) ListTools() []ToolDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	defs := make([]ToolDefinition, 0, len(s.tools))
	for _, entry := range s.tools {
		defs = append(defs, entry.Definition)
	}
	return defs
}

// ExecuteTool runs a registered tool by name with the given parameters.
func (s *Server) ExecuteTool(ctx context.Context, name string, params json.RawMessage) (any, error) {
	s.mu.RLock()
	entry, ok := s.tools[name]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("mcp: tool '%s' not found", name)
	}
	return entry.Handler(ctx, params)
}

// Start begins the MCP server's HTTP listener.
func (s *Server) Start() error {
	if !s.cfg.Enabled {
		log.Println("[MCP] Server disabled (mcp.enabled = false)")
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp/v1/tools", s.handleListTools)
	mux.HandleFunc("/mcp/v1/tools/", s.handleExecuteTool)
	mux.HandleFunc("/mcp/v1/sse", s.handleSSE)

	addr := fmt.Sprintf(":%d", s.cfg.Port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: corsMiddleware(mux),
	}

	log.Printf("[MCP] Server starting on %s (transport: %s)", addr, s.cfg.Transport)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[MCP] Server error: %v", err)
		}
	}()
	return nil
}

// Shutdown gracefully stops the MCP server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// ToolCount returns the number of registered tools.
func (s *Server) ToolCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tools)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
