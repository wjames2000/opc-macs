package runtime

import (
	"context"
	"encoding/json"
)

type PluginInfo struct {
	Name         string
	Summary      string
	Description  string
	Version      string
	Tags         []string
	Capabilities []string
	Category     string
	Parameters   map[string]interface{}
	RequiresHITL bool
	ModelName    string // 可选：插件使用的模型名，为空则用系统默认
}

type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	ModelName    string // 审计：实际调用的模型名称
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
	Trace        string // 模型思考过程追踪
}

// ModelClient 模型调用接口，每个插件通过 opts 中的 "model_client" 获取
type ModelClient interface {
	Call(ctx context.Context, req ModelRequest) (*ModelResponse, error)
	Embed(ctx context.Context, req EmbedRequest) (*EmbedResponse, error)
}

type ModelRequest struct {
	Model        string
	SystemPrompt string
	UserMessage  string
}

type ModelResponse struct {
	Content      string
	InputTokens  int
	OutputTokens int
	RawResponse  string // 模型原始响应（含思考过程）
}

type EmbedRequest struct {
	Model string
	Input string
}

type EmbedResponse struct {
	Embedding   []float32
	InputTokens int
}

// ExecuteRequest is a structured request for plugin execution.
type ExecuteRequest struct {
	Parameters json.RawMessage `json:"parameters"`
	Input      string          `json:"input"`
	Context    map[string]any  `json:"context,omitempty"`
}

// ExecuteResponse is a structured response from plugin execution.
type ExecuteResponse struct {
	Plugin  string         `json:"plugin"`
	Status  string         `json:"status"`
	Message string         `json:"message,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// ReviewRequest is a structured request for plugin review.
type ReviewRequest struct {
	Plugin     string          `json:"plugin"`
	Parameters json.RawMessage `json:"parameters"`
	Result     map[string]any  `json:"result"`
}

// ReviewResponse is a structured response from plugin review.
type ReviewResponse struct {
	Passed  bool          `json:"passed"`
	Score   float64       `json:"score"`
	Message string        `json:"message,omitempty"`
	Issues  []ReviewIssue `json:"issues,omitempty"`
}

// ReviewIssue represents a single issue found during review.
type ReviewIssue struct {
	Severity string `json:"severity"` // "error", "warning", "suggestion"
	Field    string `json:"field,omitempty"`
	Message  string `json:"message"`
}

type ExecutionResult struct {
	Data       interface{}
	TokenUsage TokenUsage
	RawTrace   string // 模型原始思考过程
}

type AgentPlugin interface {
	Name() string
	Info() PluginInfo
	Execute(ctx context.Context, input string, opts map[string]interface{}) (*ExecutionResult, error)
	Review(ctx context.Context, output interface{}) (*ReviewResult, error)
}

const PluginSymbol = "Agent"
