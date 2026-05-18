package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wjames2000/opc-macs/internal/workflow"
)

// WorkflowToolSet registers YAML workflow files as MCP tools for execution.
// This allows workflows to be triggered via MCP rather than only through the CLI.
type WorkflowToolSet struct {
	engine      *workflow.Engine
	workflowDir string
}

// NewWorkflowToolSet creates a new WorkflowToolSet.
func NewWorkflowToolSet(engine *workflow.Engine, workflowDir string) *WorkflowToolSet {
	return &WorkflowToolSet{
		engine:      engine,
		workflowDir: workflowDir,
	}
}

// RegisterAll discovers all .yaml workflow files and registers them as tools.
func (wts *WorkflowToolSet) RegisterAll(s *Server) {
	entries, err := os.ReadDir(wts.workflowDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml")) {
			continue
		}

		workflowPath := filepath.Join(wts.workflowDir, entry.Name())
		workflowName := strings.TrimSuffix(entry.Name(), ".yaml")
		workflowName = strings.TrimSuffix(workflowName, ".yml")

		wts.registerWorkflow(s, workflowName, workflowPath)
	}
}

func (wts *WorkflowToolSet) registerWorkflow(s *Server, name, path string) {
	s.RegisterTool(
		"workflow_"+name,
		fmt.Sprintf("Execute the '%s' workflow. Workflow file: %s", name, path),
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{
					"type":        "string",
					"description": "Initial input/parameters for the workflow",
				},
			},
			"required": []string{"input"},
		},
		func(ctx context.Context, params json.RawMessage) (any, error) {
			var req struct {
				Input string `json:"input"`
			}
			if err := json.Unmarshal(params, &req); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			session, err := wts.engine.StartFromFile(ctx, path, req.Input)
			if err != nil {
				return nil, fmt.Errorf("workflow start failed: %w", err)
			}

			output, _ := session.GetVar("_result")
			if output == nil {
				output, _ = session.GetVar("output")
			}
			return map[string]any{
				"workflow": name,
				"session":  session.ID,
				"status":   session.Status,
				"output":   output,
			}, nil
		},
	)
}
