package runtime

import "context"

type PluginInfo struct {
	Name         string
	Summary      string
	Version      string
	Tags         []string
	RequiresHITL bool
	ModelName    string // 可选：插件使用的模型名，为空则用系统默认
}

type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	ModelName    string // 审计：实际调用的模型名称
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

// ModelClient 模型调用接口，每个插件通过 opts 中的 "model_client" 获取
type ModelClient interface {
	Call(ctx context.Context, req ModelRequest) (*ModelResponse, error)
}

type ModelRequest struct {
	Model        string
	SystemPrompt string
	UserMessage  string
}

type ModelResponse struct {
	Content     string
	InputTokens  int
	OutputTokens int
}

type AgentPlugin interface {
	Name() string
	Info() PluginInfo
	Execute(ctx context.Context, input string, opts map[string]interface{}) (*ExecutionResult, error)
	Review(ctx context.Context, output interface{}) (*ReviewResult, error)
}

const PluginSymbol = "Agent"
