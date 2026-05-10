package runtime

import "context"

type PluginInfo struct {
	Name         string
	Summary      string
	Version      string
	Tags         []string
	RequiresHITL bool
}

type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

type ExecutionResult struct {
	Data       interface{}
	TokenUsage TokenUsage
}

type CheckResult struct {
	Item   string
	Passed bool
	Detail string
}

type ReviewResult struct {
	Passed       bool
	Score        float32
	CheckResults []CheckResult
	Summary      string
	ShouldRetry  bool
}

type AgentPlugin interface {
	Name() string
	Info() PluginInfo
	Execute(ctx context.Context, input string, opts map[string]interface{}) (*ExecutionResult, error)
	Review(ctx context.Context, output interface{}) (*ReviewResult, error)
}

const PluginSymbol = "Agent"
