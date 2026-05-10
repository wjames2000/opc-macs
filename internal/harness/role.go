package harness

import (
	"fmt"
	"strings"
)

type Action struct {
	Tool   string
	Input  string
	Output string
}

type BoundaryViolationError struct {
	AgentID string
	Action  string
	Reason  string
}

func (e *BoundaryViolationError) Error() string {
	return fmt.Sprintf("boundary violation: agent=%s action=%s reason=%s", e.AgentID, e.Action, e.Reason)
}

type RoleBoundary struct {
	AgentID         string
	Responsibilities []string
	AllowedTools    []string
	ForbiddenWords  []string
	MaxSteps        int
}

func NewRoleBoundary(agentID string, responsibilities, allowedTools, forbiddenWords []string, maxSteps int) *RoleBoundary {
	return &RoleBoundary{
		AgentID:         agentID,
		Responsibilities: responsibilities,
		AllowedTools:    allowedTools,
		ForbiddenWords:  forbiddenWords,
		MaxSteps:        maxSteps,
	}
}

func (rb *RoleBoundary) ValidateAction(action Action) error {
	for _, t := range rb.AllowedTools {
		if t == action.Tool {
			return nil
		}
	}
	return &BoundaryViolationError{
		AgentID: rb.AgentID,
		Action:  action.Tool,
		Reason:  "tool not in allowed list",
	}
}

func (rb *RoleBoundary) ValidateOutput(output string) error {
	lower := strings.ToLower(output)
	for _, word := range rb.ForbiddenWords {
		if strings.Contains(lower, strings.ToLower(word)) {
			return &BoundaryViolationError{
				AgentID: rb.AgentID,
				Action:  "output_contains_forbidden_word",
				Reason:  fmt.Sprintf("output contains forbidden word: %s", word),
			}
		}
	}
	return nil
}
