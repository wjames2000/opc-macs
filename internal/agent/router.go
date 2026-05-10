package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

// PluginSource abstracts plugin discovery for testability
type PluginSource interface {
	Get(name string) (runtime.AgentPlugin, bool)
	List() []runtime.PluginInfo
	Count() int
}

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
	plugins PluginSource
	model   string
}

func NewRouter(plugins PluginSource, model string) *Router {
	return &Router{plugins: plugins, model: model}
}

func (r *Router) buildL1Context() string {
	plugins := r.plugins.List()
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
	plugins := r.plugins.List()
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

func extractFirstJSON(raw string) string {
	start := strings.Index(raw, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(raw); i++ {
		switch raw[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return raw[start : i+1]
			}
		}
	}
	return ""
}

func parseClassifyResponse(raw string) (string, float32, error) {
	matches := extractFirstJSON(raw)
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

	// 1. Try LLM-based classification
	agentName, confidence, err := r.classify(ctx, input)
	if err != nil || agentName == "unknown" || confidence < 0.4 {
		// 2. Fallback: keyword-based classification when LLM unavailable
		agentName, confidence = r.fallbackClassify(input)
	}

	if agentName == "unknown" || confidence < 0.4 {
		return &RouteResult{
			Action:  RouteActionUnknown,
			Message: fmt.Sprintf("无法识别意图。可用 Agent：%s", r.buildAvailableList()),
		}, nil
	}

	plugin, ok := r.plugins.Get(agentName)
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

// fallbackClassify uses keyword matching when LLM is unavailable.
// Builds keyword maps from plugin names, summaries, tags, and Chinese word map.
func (r *Router) fallbackClassify(input string) (string, float32) {
	plugins := r.plugins.List()
	if len(plugins) == 0 {
		return "unknown", 0
	}

	inputLower := strings.ToLower(input)

	type match struct {
		name  string
		score int
	}
	var best, second match

	for _, p := range plugins {
		score := 0
		keywords := r.buildKeywords(p)

		for _, kw := range keywords {
			if strings.Contains(inputLower, strings.ToLower(kw)) {
				score++
			}
		}

		if score > best.score {
			second = best
			best = match{p.Name, score}
		} else if score > second.score {
			second = match{p.Name, score}
		}
	}

	// Require at least 1 match
	if best.score == 0 {
		return "unknown", 0
	}

	// If best is clearly ahead of second, high confidence
	confidence := float32(best.score) / 3.0
	if best.score > second.score+1 {
		confidence = float32(best.score) / 2.0
	}
	if confidence > 1.0 {
		confidence = 1.0
	}

	return best.name, confidence
}

// chineseKeywords maps agent names to Chinese keyword patterns for
// fallback classification when LLM is unavailable.
var chineseKeywords = map[string][]string{
	"copywriter":   {"文案", "推广", "营销", "广告", "宣传", "产品", "描述", "标题", "广告词", "促销"},
	"email_sorter": {"邮件", "email", "投诉", "退款", "咨询", "客户", "回复", "来信", "收件", "发件"},
	"xhs_poster":   {"小红书", "种草", "笔记", "xhs", "好物", "测评", "推荐", "安利", "分享"},
}

// buildKeywords generates search keywords from a plugin's metadata and Chinese keywords
func (r *Router) buildKeywords(p runtime.PluginInfo) []string {
	keywords := make([]string, 0)
	seen := make(map[string]bool)

	add := func(kw string) {
		lower := strings.ToLower(strings.TrimSpace(kw))
		if lower != "" && !seen[lower] {
			seen[lower] = true
			keywords = append(keywords, lower)
		}
	}

	// From name
	add(p.Name)
	// From summary - add as single string for Chinese
	add(p.Summary)
	// From tags
	for _, tag := range p.Tags {
		add(tag)
	}
	// From Chinese keyword map
	if ckw, ok := chineseKeywords[p.Name]; ok {
		for _, kw := range ckw {
			add(kw)
		}
	}

	return keywords
}

func (r *Router) buildAvailableList() string {
	plugins := r.plugins.List()
	names := make([]string, len(plugins))
	for i, p := range plugins {
		names[i] = p.Name
	}
	return strings.Join(names, ", ")
}
