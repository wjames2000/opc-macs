// Package plugins provides embedded agent implementations for development mode.
// In production (Linux), these same agents can be compiled as .so plugins.
package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

func RegisterAll(loader *runtime.Loader) error {
	agents := []runtime.AgentPlugin{
		&CopywriterPlugin{},
		&EmailSorterPlugin{},
		&XHSPosterPlugin{},
		&CompetitiveAnalysisPlugin{},
		&MeetingMinutesPlugin{},
	}
	for _, a := range agents {
		if err := loader.Register(a); err != nil {
			return err
		}
	}
	return nil
}

// ──────────────────────────────────────────────
// CopywriterPlugin — 营销文案生成
// ──────────────────────────────────────────────

var copywriterSystemPrompt = `你是专业的营销文案撰写专家。你的任务是根据产品信息生成三段式营销文案。

[角色边界]
- 只做文案生成，不做市场分析
- 不使用夸张/虚假宣传词汇
- 风格根据产品类型自动适配

[输出要求]
- short_copy: 标题级短文案，≤20字，有冲击力
- long_copy: 详情文案，100-200字，含产品卖点
- social_copy: 社交媒体文案，≤80字，适合朋友圈/推文
- style: 文案风格，从[科技简约, 温暖亲和, 专业正式, 年轻活力]中选择

请以 JSON 格式输出，不要包含其他内容。`

type CopywriterPlugin struct{}

func (p *CopywriterPlugin) Name() string { return "copywriter" }

func (p *CopywriterPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "copywriter",
		Summary:      "根据产品信息生成三段式营销文案",
		Version:      "1.0.0",
		Tags:         []string{"marketing", "copywriting", "social-media"},
		RequiresHITL: true,
		ModelName:    "gemini-2.0-flash",
	}
}

func (p *CopywriterPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("copywriter: 未配置模型 API，请在 config.yaml 中设置 model.api_key")
	}
	modelName := resolveModelName(opts, p.Info().ModelName)

	memories, _ := opts["memories"].([]string)
	history, _ := opts["history"].(string)
	var sb strings.Builder
	sb.WriteString(input)
	if len(memories) > 0 {
		sb.WriteString("\n\n参考历史信息：\n")
		for i, m := range memories {
			if i >= 3 {
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", m))
		}
	}
	if history != "" {
		sb.WriteString("\n\n")
		sb.WriteString(history)
	}

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: copywriterSystemPrompt,
		UserMessage:  sb.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("copywriter: 模型调用失败：%w", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		return nil, fmt.Errorf("copywriter: 模型输出解析失败（期望 JSON 格式）：%w", err)
	}

	return &runtime.ExecutionResult{
		Data: output,
		TokenUsage: runtime.TokenUsage{
			InputTokens:  resp.InputTokens,
			OutputTokens: resp.OutputTokens,
			ModelName:    modelName,
		},
		RawTrace: resp.RawResponse,
	}, nil
}

func (p *CopywriterPlugin) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output"}, nil
	}
	var checks []runtime.CheckResult
	allPassed := true
	if s, ok := data["short_copy"].(string); ok {
		passed := len(s) <= 20
		if !passed {
			allPassed = false
		}
		checks = append(checks, runtime.CheckResult{Item: "short_copy ≤ 20", Passed: passed})
	}
	if s, ok := data["long_copy"].(string); ok {
		passed := len(s) >= 50 && len(s) <= 500
		if !passed {
			allPassed = false
		}
		checks = append(checks, runtime.CheckResult{Item: "long_copy 50-500", Passed: passed})
	}
	if s, ok := data["social_copy"].(string); ok {
		passed := len(s) <= 80
		if !passed {
			allPassed = false
		}
		checks = append(checks, runtime.CheckResult{Item: "social_copy ≤ 80", Passed: passed})
	}
	return &runtime.ReviewResult{
		Passed: allPassed, Score: scoreFromChecks(checks),
		CheckResults: checks, ShouldRetry: !allPassed,
		Summary: fmt.Sprintf("copywriter: %d/%d checks passed", countPassed(checks), len(checks)),
	}, nil
}

// ──────────────────────────────────────────────
// EmailSorterPlugin — 邮件分类
// ──────────────────────────────────────────────

var emailSorterSystemPrompt = `你是专业的客服邮件处理专家。你的任务是对客户邮件进行分类并提供回复建议。

[角色边界]
- 只做邮件分类和回复建议
- 不做实际发送操作
- 紧急邮件优先标记

[分类标准]
- 咨询: 询问产品/服务信息，语气正常
- 投诉: 表达不满，要求退款/赔偿
- 合作: 商务合作/推广邀约
- 垃圾: 广告/诈骗/无关内容

[输出要求]
- category: 从[咨询, 投诉, 合作, 垃圾]中选择
- reason: 分类理由，20-50字
- reply_suggestion: 回复草稿，50-200字，专业礼貌
- urgency: 从[低, 中, 高]中选择

请以 JSON 格式输出。`

type EmailSorterPlugin struct{}

func (p *EmailSorterPlugin) Name() string { return "email_sorter" }

func (p *EmailSorterPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "email_sorter",
		Summary:      "分析邮件内容并分类，生成回复建议",
		Version:      "1.0.0",
		Tags:         []string{"email", "customer-service", "classification"},
		RequiresHITL: true,
		ModelName:    "gemini-2.0-flash",
	}
}

func (p *EmailSorterPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("email_sorter: 未配置模型 API，请在 config.yaml 中设置 model.api_key")
	}
	modelName := resolveModelName(opts, p.Info().ModelName)

	// Spam keyword pre-check (bypasses LLM)
	lower := strings.ToLower(input)
	spamWords := []string{"免费", "中奖", "转账", "点击链接", "free", "winner", "bank transfer"}
	for _, w := range spamWords {
		if strings.Contains(lower, w) {
			return &runtime.ExecutionResult{
				Data: map[string]interface{}{
					"category": "垃圾", "reason": "命中垃圾关键词规则",
					"reply_suggestion": "", "urgency": "低",
				},
				TokenUsage: runtime.TokenUsage{ModelName: "rule-based"},
			}, nil
		}
	}

	payload := input
	if h, _ := opts["history"].(string); h != "" {
		payload = h + "\n\n当前邮件：\n" + input
	}

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: emailSorterSystemPrompt,
		UserMessage:  payload,
	})
	if err != nil {
		return nil, fmt.Errorf("email_sorter: 模型调用失败：%w", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		return nil, fmt.Errorf("email_sorter: 模型输出解析失败（期望 JSON 格式）：%w", err)
	}

	return &runtime.ExecutionResult{
		Data: output,
		TokenUsage: runtime.TokenUsage{
			InputTokens:  resp.InputTokens,
			OutputTokens: resp.OutputTokens,
			ModelName:    modelName,
		},
		RawTrace: resp.RawResponse,
	}, nil
}

func (p *EmailSorterPlugin) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output"}, nil
	}
	cat, _ := data["category"].(string)
	valid := map[string]bool{"咨询": true, "投诉": true, "合作": true, "垃圾": true}
	passed := valid[cat]
	return &runtime.ReviewResult{
		Passed:      passed,
		Score:       boolToScore(passed),
		ShouldRetry: !passed,
		Summary:     fmt.Sprintf("email_sorter: category=%s valid=%v", cat, passed),
	}, nil
}

// ──────────────────────────────────────────────
// XHSPosterPlugin — 小红书种草笔记
// ──────────────────────────────────────────────

var xhsSystemPrompt = `你是资深小红书内容创作者，擅长用亲切自然的语气撰写种草笔记。

[角色边界]
- 只做内容生成，不做竞品分析
- 不使用虚假夸张宣传
- 风格必须符合小红书社区规范
- 适当使用emoji增加可读性

[输出要求]
- title: 笔记标题，≤20字，含emoji，吸引眼球
- body: 笔记正文，200-500字，使用小红书风格用语（"姐妹们"/"亲测"/"安利"），含emoji分段
- hashtags: 5-10个话题标签，如 #好物推荐 #种草
- image_suggestions: 2-3张配图描述建议
- style: 从[好物推荐, 使用心得, 开箱测评, 生活记录, 教程攻略]中选择

请以 JSON 格式输出。`

type XHSPosterPlugin struct{}

func (p *XHSPosterPlugin) Name() string { return "xhs_poster" }

func (p *XHSPosterPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "xhs_poster",
		Summary:      "生成小红书种草笔记内容",
		Version:      "1.0.0",
		Tags:         []string{"social-media", "xiaohongshu", "content-marketing"},
		RequiresHITL: true,
		ModelName:    "gemini-2.0-flash",
	}
}

func (p *XHSPosterPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("xhs_poster: 未配置模型 API，请在 config.yaml 中设置 model.api_key")
	}
	modelName := resolveModelName(opts, p.Info().ModelName)

	memories, _ := opts["memories"].([]string)
	history, _ := opts["history"].(string)
	var sb strings.Builder
	sb.WriteString(input)
	if len(memories) > 0 {
		sb.WriteString("\n\n参考历史内容：\n")
		for i, m := range memories {
			if i >= 3 {
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", m))
		}
	}
	if history != "" {
		sb.WriteString("\n\n")
		sb.WriteString(history)
	}

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: xhsSystemPrompt,
		UserMessage:  sb.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("xhs_poster: 模型调用失败：%w", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		return nil, fmt.Errorf("xhs_poster: 模型输出解析失败（期望 JSON 格式）：%w", err)
	}

	return &runtime.ExecutionResult{
		Data: output,
		TokenUsage: runtime.TokenUsage{
			InputTokens:  resp.InputTokens,
			OutputTokens: resp.OutputTokens,
			ModelName:    modelName,
		},
		RawTrace: resp.RawResponse,
	}, nil
}

func (p *XHSPosterPlugin) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output"}, nil
	}
	var checks []runtime.CheckResult
	if t, ok := data["title"].(string); ok {
		checks = append(checks, runtime.CheckResult{Item: "title ≤ 20", Passed: len(t) <= 20})
	}
	if b, ok := data["body"].(string); ok {
		checks = append(checks, runtime.CheckResult{Item: "body 200-500", Passed: len(b) >= 200 && len(b) <= 500})
	}
	if h, ok := data["hashtags"].([]interface{}); ok {
		checks = append(checks, runtime.CheckResult{Item: "hashtags 5-10", Passed: len(h) >= 5 && len(h) <= 10})
	}
	allPassed := true
	for _, c := range checks {
		if !c.Passed {
			allPassed = false
			break
		}
	}
	return &runtime.ReviewResult{
		Passed: allPassed, Score: scoreFromChecks(checks),
		CheckResults: checks, ShouldRetry: !allPassed,
		Summary: fmt.Sprintf("xhs_poster: %d/%d checks passed", countPassed(checks), len(checks)),
	}, nil
}

// ──────────────────────────────────────────────
// CompetitiveAnalysisPlugin — 竞品分析
// ──────────────────────────────────────────────

var competitiveSystemPrompt = `你是资深的商业分析师，擅长竞品分析。请对用户提供的竞品信息进行结构化分析。

[输出格式要求 JSON]
{
  "competitor": "竞品名称",
  "market_position": "市场定位概述",
  "strengths": ["优势1", "优势2", "优势3"],
  "weaknesses": ["劣势1", "劣势2", "劣势3"],
  "opportunities": ["机会1", "机会2"],
  "threats": ["威胁1", "威胁2"],
  "differentiation": "差异化建议",
  "risk_level": "低/中/高",
  "summary": "综合结论"
}`

type CompetitiveAnalysisPlugin struct{}

func (p *CompetitiveAnalysisPlugin) Name() string { return "competitive_analysis" }

func (p *CompetitiveAnalysisPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "competitive_analysis",
		Summary:      "对竞品进行 SWOT 结构化分析",
		Version:      "1.0.0",
		Tags:         []string{"analysis", "strategy", "business", "SWOT"},
		RequiresHITL: false,
	}
}

func (p *CompetitiveAnalysisPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("competitive_analysis: 未配置模型 API")
	}
	modelName := resolveModelName(opts, p.Info().ModelName)

	memories, _ := opts["memories"].([]string)
	var sb strings.Builder
	sb.WriteString(input)
	if len(memories) > 0 {
		sb.WriteString("\n\n参考历史信息：\n")
		for i, m := range memories {
			if i >= 3 {
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", m))
		}
	}
	if h, _ := opts["history"].(string); h != "" {
		sb.WriteString("\n\n")
		sb.WriteString(h)
	}

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: competitiveSystemPrompt,
		UserMessage:  sb.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("competitive_analysis: 模型调用失败：%w", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		return nil, fmt.Errorf("competitive_analysis: 输出解析失败：%w", err)
	}

	return &runtime.ExecutionResult{
		Data: output,
		TokenUsage: runtime.TokenUsage{InputTokens: resp.InputTokens, OutputTokens: resp.OutputTokens, ModelName: modelName},
		RawTrace:   resp.RawResponse,
	}, nil
}

func (p *CompetitiveAnalysisPlugin) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output"}, nil
	}
	checks := []runtime.CheckResult{}
	if _, ok := data["strengths"].([]interface{}); ok {
		checks = append(checks, runtime.CheckResult{Item: "strengths present", Passed: true})
	}
	if _, ok := data["weaknesses"].([]interface{}); ok {
		checks = append(checks, runtime.CheckResult{Item: "weaknesses present", Passed: true})
	}
	if _, ok := data["risk_level"].(string); ok {
		checks = append(checks, runtime.CheckResult{Item: "risk_level present", Passed: true})
	}
	return &runtime.ReviewResult{
		Passed: true, Score: scoreFromChecks(checks),
		CheckResults: checks, Summary: "competitive_analysis: analysis complete",
	}, nil
}

// ──────────────────────────────────────────────
// MeetingMinutesPlugin — 会议纪要
// ──────────────────────────────────────────────

var meetingMinutesSystemPrompt = `你是专业的会议纪要撰写助手。请将会议讨论内容整理为结构化的会议纪要。

[输出格式要求 JSON]
{
  "title": "会议主题",
  "time": "会议时间（从内容推断）",
  "participants": ["参会人1", "参会人2"],
  "agenda": ["议题1", "议题2"],
  "decisions": ["决策1", "决策2"],
  "action_items": [
    {"task": "待办事项", "owner": "负责人", "deadline": "截止日期"}
  ],
  "next_steps": "下一步计划",
  "key_discussions": "关键讨论内容摘要"
}`

type MeetingMinutesPlugin struct{}

func (p *MeetingMinutesPlugin) Name() string { return "meeting_minutes" }

func (p *MeetingMinutesPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "meeting_minutes",
		Summary:      "将会议讨论整理为结构化会议纪要",
		Version:      "1.0.0",
		Tags:         []string{"meeting", "minutes", "productivity", "document"},
		RequiresHITL: false,
	}
}

func (p *MeetingMinutesPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("meeting_minutes: 未配置模型 API")
	}
	modelName := resolveModelName(opts, p.Info().ModelName)

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: meetingMinutesSystemPrompt,
		UserMessage:  input,
	})
	if err != nil {
		return nil, fmt.Errorf("meeting_minutes: 模型调用失败：%w", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		return nil, fmt.Errorf("meeting_minutes: 输出解析失败：%w", err)
	}

	return &runtime.ExecutionResult{
		Data: output,
		TokenUsage: runtime.TokenUsage{InputTokens: resp.InputTokens, OutputTokens: resp.OutputTokens, ModelName: modelName},
		RawTrace:   resp.RawResponse,
	}, nil
}

func (p *MeetingMinutesPlugin) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output"}, nil
	}
	checks := []runtime.CheckResult{}
	if _, ok := data["decisions"].([]interface{}); ok {
		checks = append(checks, runtime.CheckResult{Item: "decisions present", Passed: true})
	}
	if _, ok := data["action_items"].([]interface{}); ok {
		checks = append(checks, runtime.CheckResult{Item: "action_items present", Passed: true})
	}
	return &runtime.ReviewResult{
		Passed: true, Score: scoreFromChecks(checks),
		CheckResults: checks, Summary: "meeting_minutes: processed",
	}, nil
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func resolveModelName(opts map[string]interface{}, pluginModel string) string {
	// Priority: opts from config > plugin-specific override > default
	if m, ok := opts["model_name"].(string); ok && m != "" {
		return m
	}
	if pluginModel != "" {
		return pluginModel
	}
	return "deepseek-v4-flash"
}

func scoreFromChecks(checks []runtime.CheckResult) float32 {
	if len(checks) == 0 {
		return 5.0
	}
	return float32(countPassed(checks)) / float32(len(checks)) * 5.0
}

func countPassed(checks []runtime.CheckResult) int {
	n := 0
	for _, c := range checks {
		if c.Passed {
			n++
		}
	}
	return n
}

func boolToScore(b bool) float32 {
	if b {
		return 5.0
	}
	return 0
}
