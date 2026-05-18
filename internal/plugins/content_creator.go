package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

var contentCreatorSystemPrompt = `你是全能内容创作专家，使用 AI 为社交媒体生成图文、视频、文案等多种形式的内容。

[核心能力]
- 文案创作：营销文案、种草笔记、短视频脚本、长文深度文章
- 图片生成：基于描述生成配图，支持多风格（摄影/插画/3D/设计）
- 视频脚本：分镜脚本、话术设计、镜头描述
- 内容扩写/缩写：短内容扩写成长文，长文缩写成摘要

[创作风格]
- 带货：产品卖点突出、促单话术、限时感
- 知识：专业深度、数据支撑、逻辑清晰
- 娱乐：幽默有趣、互动性强、梗和热梗
- 教程：步骤清晰、易于跟学、配图说明
- 测评：客观公正、多维度对比、真实体验
- Vlog：真实自然、故事性、代入感
- 剧情：剧本结构、冲突设置、悬念

[平台适配]
不同平台内容风格自动调整：
- 抖音/快手：短视频文案，标题党，快节奏
- B站：中长视频，深度内容，专业调性
- 小红书：图文种草，真实体验，视觉突出
- 微信：长文深度，情感共鸣
- YouTube：高质量长视频，信息密度高
- TikTok/IG：国际化，视觉驱动，简短有力

[输出要求]
以 JSON 格式输出创作内容，包含内容类型、标题、正文/文案、配图描述、标签建议和适用平台列表。`

type ContentCreatorPlugin struct{}

func (p *ContentCreatorPlugin) Name() string { return "content_creator" }

func (p *ContentCreatorPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "content_creator",
		Summary:      "AI 内容创作：文案/图片/视频脚本/多风格多平台内容生成",
		Version:      "1.0.0",
		Tags:         []string{"create", "content", "copywriting", "video-script", "image-generation"},
		RequiresHITL: true,
		ModelName:    "gemini-2.0-flash",
	}
}

func (p *ContentCreatorPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("content_creator: model client not configured")
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
		SystemPrompt: contentCreatorSystemPrompt,
		UserMessage:  sb.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("content_creator: model call failed: %w", err)
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

func (p *ContentCreatorPlugin) Review(ctx context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output format"}, nil
	}

	var checks []runtime.CheckResult
	allPassed := true

	contentTypes := []string{"title", "body", "content_type"}
	for _, field := range contentTypes {
		if val, ok := data[field].(string); ok && val != "" {
			checks = append(checks, runtime.CheckResult{Item: field + " present", Passed: true})
		}
	}

	if platforms, ok := data["platforms"].([]interface{}); ok && len(platforms) > 0 {
		checks = append(checks, runtime.CheckResult{Item: "platforms specified", Passed: true, Detail: fmt.Sprintf("%d platforms", len(platforms))})
	}

	return &runtime.ReviewResult{
		Passed:       allPassed,
		Score:        calculateScore(checks),
		CheckResults: checks,
		Summary:      summaryStr(checks),
		ShouldRetry:  !allPassed,
	}, nil
}
