package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wjames2000/opc-macs/internal/agent"
	"github.com/wjames2000/opc-macs/internal/runtime"
)

// AgentToolSet registers all available agent plugins as MCP tools.
// This allows Claude and other AI assistants to invoke agents directly via MCP.
type AgentToolSet struct {
	router      *agent.Router
	loader      *runtime.Loader
	modelClient runtime.ModelClient
	modelName   string
}

// NewAgentToolSet creates a new AgentToolSet.
func NewAgentToolSet(router *agent.Router, loader *runtime.Loader, modelClient runtime.ModelClient, modelName string) *AgentToolSet {
	return &AgentToolSet{
		router:      router,
		loader:      loader,
		modelClient: modelClient,
		modelName:   modelName,
	}
}

// RegisterAll registers all available agents as MCP tools.
func (ats *AgentToolSet) RegisterAll(s *Server) {
	for _, info := range ats.loader.List() {
		plugin, ok := ats.loader.Get(info.Name)
		if !ok {
			continue
		}
		ats.registerAgentAsTool(s, plugin, info)
	}
}

func (ats *AgentToolSet) registerAgentAsTool(s *Server, plugin runtime.AgentPlugin, info runtime.PluginInfo) {
	// Build input schema from tags for basic validation
	properties := map[string]any{
		"input": map[string]any{
			"type":        "string",
			"description": fmt.Sprintf("Task description for %s agent: %s", info.Name, info.Summary),
		},
	}
	if info.RequiresHITL {
		properties["confirm_before_execute"] = map[string]any{
			"type":        "boolean",
			"description": "Set to false to require human confirmation before executing",
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   []string{"input"},
	}

	s.RegisterTool(
		"agent_"+info.Name,
		fmt.Sprintf("[Agent] %s: %s (tags: %v)", info.Name, info.Summary, info.Tags),
		schema,
		func(ctx context.Context, params json.RawMessage) (any, error) {
			var req struct {
				Input string `json:"input"`
			}
			if err := json.Unmarshal(params, &req); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			opts := map[string]interface{}{
				"model_client": ats.modelClient,
			}

			result, err := plugin.Execute(ctx, req.Input, opts)
			if err != nil {
				return nil, fmt.Errorf("agent %s execute failed: %w", info.Name, err)
			}

			return map[string]any{
				"agent":       info.Name,
				"result":      result.Data,
				"token_usage": result.TokenUsage,
			}, nil
		},
	)
}

// AgentRouterTool registers a tool that uses the LLM router to select and run the right agent.
func (ats *AgentToolSet) RegisterRouterTool(s *Server) {
	s.RegisterTool(
		"run_agent",
		"Route a task description to the most suitable agent and execute it. The system automatically selects the best agent based on the task description.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"task": map[string]any{
					"type":        "string",
					"description": "The task description. Be as specific as possible about what you want the agent to do.",
				},
			},
			"required": []string{"task"},
		},
		func(ctx context.Context, params json.RawMessage) (any, error) {
			var req struct {
				Task string `json:"task"`
			}
			if err := json.Unmarshal(params, &req); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			routeResult, routeErr := ats.router.Route(ctx, req.Task)
			if routeErr != nil || routeResult.Action == agent.RouteActionUnknown {
				return map[string]any{
					"error":   "no suitable agent found",
					"message": "The task could not be matched to any available agent. Try being more specific or using a different tool.",
				}, nil
			}

			opts := map[string]interface{}{
				"model_client": ats.modelClient,
			}

			execResult, err := routeResult.Plugin.Execute(ctx, req.Task, opts)
			if err != nil {
				return nil, fmt.Errorf("agent execution failed: %w", err)
			}

			return map[string]any{
				"agent":       routeResult.Info.Name,
				"summary":     routeResult.Info.Summary,
				"result":      execResult.Data,
				"token_usage": execResult.TokenUsage,
			}, nil
		},
	)
}
