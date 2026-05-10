package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type RouteAction int

const (
	RouteActionDispatch RouteAction = iota
	RouteActionUnknown
)

type RouteResult struct {
	Action  RouteAction
	Plugin  runtime.AgentPlugin
	Info    runtime.PluginInfo
	Message string
}

type classifyResponse struct {
	Agent      string  `json:"agent"`
	Confidence float32 `json:"confidence"`
	Reason     string  `json:"reason"`
}

type Router struct {
	loader *runtime.Loader
	model  string
}

func NewRouter(loader *runtime.Loader, model string) *Router {
	return &Router{loader: loader, model: model}
}

func (r *Router) buildL1Context() string {
	plugins := r.loader.List()
	var sb strings.Builder
	sb.WriteString("你是一个路由智能体，根据用户输入判断应该分发到哪个 Agent。\n\n")
	sb.WriteString("可用 Agent：\n")
	for _, p := range plugins {
		sb.WriteString(fmt.Sprintf("- %s：%s\n", p.Name, p.Summary))
	}
	sb.WriteString("\n如果无法匹配任何 Agent，返回 unknown。")
	return sb.String()
}

const classifyPromptTemplate = `
你是 OPC-Agent 的路由智能体。
请判断以下用户输入属于哪个 Agent 的技能范围。

可用 Agent：
{{plugins}}

用户输入：{{input}}

请以 JSON 格式返回（仅返回 JSON）：
{
    "agent": "Agent 名称 或 unknown",
    "confidence": 0.0-1.0,
    "reason": "简短分类理由"
}
`

func (r *Router) buildClassifyPrompt(input string) string {
	plugins := r.loader.List()
	var pluginLines strings.Builder
	for _, p := range plugins {
		pluginLines.WriteString(fmt.Sprintf("- %s：%s\n", p.Name, p.Summary))
	}
	prompt := strings.ReplaceAll(classifyPromptTemplate, "{{plugins}}", pluginLines.String())
	prompt = strings.ReplaceAll(prompt, "{{input}}", input)
	return prompt
}

func (r *Router) classify(ctx context.Context, input string) (string, float32, error) {
	prompt := r.buildClassifyPrompt(input)

	response, err := callLLM(ctx, r.model, prompt)
	if err != nil {
		return "", 0, fmt.Errorf("router: llm call failed: %w", err)
	}

	return parseClassifyResponse(response)
}

var jsonExtractRe = regexp.MustCompile(`\{[^{}]*\}`)

func parseClassifyResponse(raw string) (string, float32, error) {
	matches := jsonExtractRe.FindString(raw)
	if matches == "" {
		return "", 0, fmt.Errorf("router: no JSON found in response")
	}

	var resp classifyResponse
	if err := json.Unmarshal([]byte(matches), &resp); err != nil {
		return "", 0, fmt.Errorf("router: json parse failed: %w", err)
	}

	return resp.Agent, resp.Confidence, nil
}

func (r *Router) Route(ctx context.Context, input string) (*RouteResult, error) {
	if strings.TrimSpace(input) == "" {
		return &RouteResult{
			Action:  RouteActionUnknown,
			Message: "请输入任务描述。",
		}, nil
	}

	agentName, confidence, err := r.classify(ctx, input)
	if err != nil {
		return &RouteResult{
			Action:  RouteActionUnknown,
			Message: "识别失败，请重新描述。",
		}, nil
	}

	if agentName == "unknown" || confidence < 0.4 {
		return &RouteResult{
			Action:  RouteActionUnknown,
			Message: fmt.Sprintf("无法识别意图。可用 Agent：%s", r.buildAvailableList()),
		}, nil
	}

	plugin, ok := r.loader.Get(agentName)
	if !ok {
		return &RouteResult{
			Action:  RouteActionUnknown,
			Message: fmt.Sprintf("Agent '%s' 未加载", agentName),
		}, nil
	}

	return &RouteResult{
		Action: RouteActionDispatch,
		Plugin: plugin,
		Info:   plugin.Info(),
	}, nil
}

func (r *Router) buildAvailableList() string {
	plugins := r.loader.List()
	names := make([]string, len(plugins))
	for i, p := range plugins {
		names[i] = p.Name
	}
	return strings.Join(names, ", ")
}
