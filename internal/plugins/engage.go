package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

var engageSystemPrompt = `你是社交媒体互动运营专家，负责管理各大平台的评论回复、互动分析和用户参与。

[核心能力]
- 评论回复：自动回复平台评论，支持个性化话术
- 互动管理：点赞、置顶、删除评论
- 互动分析：分析评论情感、关键词提取、舆情监控
- 批量运营：批量回复相似评论，自动处理常见问题

[支持平台]
- 抖音、小红书、B站、快手、微信视频号
- YouTube、TikTok、Facebook、Instagram、Twitter

[互动策略]
- 正面评论：感谢+互动引导
- 中性评论：解答疑问+补充信息
- 负面评论：先安抚后解决，避免正面冲突
- 常见问题：统一话术模板
- 深度互动：针对优质评论展开讨论

[输出要求]
以 JSON 格式输出互动计划，包含目标平台、互动类型、回复内容和时间安排。`

// EngagePlugin implements the engage Agent for social media interaction management.
type EngagePlugin struct{}

func (p *EngagePlugin) Name() string { return "engage" }

func (p *EngagePlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "engage",
		Summary:      "社交媒体互动运营：评论回复、互动管理、舆情分析",
		Version:      "1.0.0",
		Tags:         []string{"engage", "social-media", "comment", "interaction"},
		RequiresHITL: true,
		ModelName:    "gemini-2.0-flash",
	}
}

func (p *EngagePlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("engage: model client not configured")
	}
	modelName := resolveModelName(opts, p.Info().ModelName)

	memories, _ := opts["memories"].([]string)
	var sb strings.Builder
	sb.WriteString(input)
	if len(memories) > 0 {
		sb.WriteString("\n\n参考信息：\n")
		for i, m := range memories {
			if i >= 3 {
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", m))
		}
	}

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: engageSystemPrompt,
		UserMessage:  sb.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("engage: model call failed: %w", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		output = map[string]interface{}{
			"raw_response": resp.Content,
			"note":         "Response was not valid JSON, returned as raw text",
		}
	}

	return &runtime.ExecutionResult{
		Data:       output,
		TokenUsage: runtime.TokenUsage{InputTokens: resp.InputTokens, OutputTokens: resp.OutputTokens},
	}, nil
}

func (p *EngagePlugin) Review(ctx context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output format"}, nil
	}

	var checks []runtime.CheckResult
	allPassed := true

	if platform, ok := data["platform"].(string); ok && platform != "" {
		checks = append(checks, runtime.CheckResult{Item: "platform specified", Passed: true, Detail: platform})
	} else {
		allPassed = false
		checks = append(checks, runtime.CheckResult{Item: "platform specified", Passed: false, Detail: "missing"})
	}

	if action, ok := data["action"].(string); ok && action != "" {
		checks = append(checks, runtime.CheckResult{Item: "action specified", Passed: true, Detail: action})
	} else {
		allPassed = false
		checks = append(checks, runtime.CheckResult{Item: "action specified", Passed: false, Detail: "missing"})
	}

	return &runtime.ReviewResult{
		Passed:       allPassed,
		Score:        calculateScore(checks),
		CheckResults: checks,
		Summary:      summaryStr(checks),
		ShouldRetry:  !allPassed,
	}, nil
}
