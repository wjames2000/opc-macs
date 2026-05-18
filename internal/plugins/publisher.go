package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

var publisherSystemPrompt = `你是内容发布专家，负责将用户的内容一键分发到多个社交平台。

[核心能力]
- 内容适配：自动将内容调整到各平台的格式要求（标题长度、文案长度、视频比例、标签格式）
- 多平台发布：支持同时发布到抖音、小红书、B站、快手、微信、YouTube、TikTok、Facebook、Instagram、Twitter、LinkedIn、Pinterest
- 定时发布：支持指定发布时间
- 发布队列：批量排队发布，自动重试

[平台特殊要求]
- 抖音：9:16竖屏视频，标题≤30字，文案≤1000字，最多10个标签
- 小红书：图文或视频，标题≤20字，文案≤2000字，最多10个标签
- B站：16:9横屏视频，标题≤50字，文案≤3000字，最多6个标签
- YouTube：16:9横屏视频，标题≤100字，文案≤5000字，最多15个标签
- TikTok：9:16竖屏视频，标题≤50字
- Instagram：1:1或9:16图片/视频，文案≤2200字，最多30个标签
- Twitter：文案≤280字
- LinkedIn：标题≤70字，文案≤3000字

[输出要求]
以 JSON 格式输出发布计划，包含目标平台、适配后的内容和发布时间。`

type PublisherPlugin struct{}

func (p *PublisherPlugin) Name() string { return "publisher" }

func (p *PublisherPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "publisher",
		Summary:      "一键多平台内容发布，支持 13+ 社交平台同步发布、内容格式自动适配、定时发布",
		Version:      "1.0.0",
		Tags:         []string{"publish", "social-media", "content-marketing", "multi-platform"},
		RequiresHITL: true,
		ModelName:    "gemini-2.0-flash",
	}
}

func (p *PublisherPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("publisher: model client not configured")
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
		SystemPrompt: publisherSystemPrompt,
		UserMessage:  sb.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("publisher: model call failed: %w", err)
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

func (p *PublisherPlugin) Review(ctx context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output format"}, nil
	}

	var checks []runtime.CheckResult
	allPassed := true

	if platforms, ok := data["platforms"].([]interface{}); ok {
		passed := len(platforms) > 0
		if !passed {
			allPassed = false
		}
		checks = append(checks, runtime.CheckResult{
			Item: "at least one target platform", Passed: passed,
			Detail: fmt.Sprintf("count=%d", len(platforms)),
		})
	}

	if content, ok := data["content"].(map[string]interface{}); ok {
		if title, ok := content["title"].(string); ok && title != "" {
			checks = append(checks, runtime.CheckResult{
				Item: "title present", Passed: true, Detail: fmt.Sprintf("len=%d", len(title)),
			})
		}
		if body, ok := content["body"].(string); ok && body != "" {
			checks = append(checks, runtime.CheckResult{
				Item: "body present", Passed: true, Detail: fmt.Sprintf("len=%d", len(body)),
			})
		}
	}

	return &runtime.ReviewResult{
		Passed:       allPassed,
		Score:        calculateScore(checks),
		CheckResults: checks,
		Summary:      summaryStr(checks),
		ShouldRetry:  !allPassed,
	}, nil
}
