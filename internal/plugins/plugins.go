// Package plugins provides embedded agent implementations for development mode.
// In production (Linux), these same agents can be compiled as .so plugins.
package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

// RegisterAll registers all built-in agents with the loader.
// Used when .so plugin loading is unavailable (e.g., macOS development).
func RegisterAll(loader *runtime.Loader) error {
	agents := []runtime.AgentPlugin{
		&CopywriterPlugin{},
		&EmailSorterPlugin{},
		&XHSPosterPlugin{},
	}
	for _, a := range agents {
		if err := loader.Register(a); err != nil {
			return err
		}
	}
	return nil
}

// CopywriterPlugin - embedded copywriter agent
type CopywriterPlugin struct{}

func (p *CopywriterPlugin) Name() string { return "copywriter" }

func (p *CopywriterPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "copywriter",
		Summary:      "根据产品信息生成三段式营销文案",
		Version:      "1.0.0",
		Tags:         []string{"marketing", "copywriting", "social-media"},
		RequiresHITL: true,
	}
}

func (p *CopywriterPlugin) Execute(_ context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	// MVP stub - returns sample output
	memories, _ := opts["memories"].([]string)
	_ = memories

	return &runtime.ExecutionResult{
		Data: map[string]interface{}{
			"short_copy":  fmt.Sprintf("「%s」懂你所需，品质之选", truncate(input, 8)),
			"long_copy":   fmt.Sprintf("在寻找一款真正懂你的产品吗？%s\n\n我们用心打磨每一个细节，只为给你最好的体验。从选材到工艺，每一步都精益求精。\n\n现在就行动起来，开启品质生活新篇章。", input),
			"social_copy": fmt.Sprintf("被问爆了！%s 真的绝绝子✨", truncate(input, 10)),
			"style":       "科技简约",
		},
		TokenUsage: runtime.TokenUsage{InputTokens: 100, OutputTokens: 150},
	}, nil
}

func (p *CopywriterPlugin) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output"}, nil
	}
	allPassed := true
	var checks []runtime.CheckResult
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
	return &runtime.ReviewResult{
		Passed: allPassed, Score: scoreFromChecks(checks),
		CheckResults: checks, ShouldRetry: !allPassed,
	}, nil
}

// EmailSorterPlugin - embedded email classification agent
type EmailSorterPlugin struct{}

func (p *EmailSorterPlugin) Name() string { return "email_sorter" }

func (p *EmailSorterPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name: "email_sorter", Summary: "分析邮件内容并分类，生成回复建议",
		Version: "1.0.0", Tags: []string{"email", "customer-service", "classification"},
		RequiresHITL: true,
	}
}

func (p *EmailSorterPlugin) Execute(_ context.Context, input string, _ map[string]interface{}) (*runtime.ExecutionResult, error) {
	lower := strings.ToLower(input)
	cat, urgency := "咨询", "低"
	if strings.Contains(lower, "投诉") || strings.Contains(lower, "退款") || strings.Contains(lower, "complaint") {
		cat, urgency = "投诉", "高"
	} else if strings.Contains(lower, "合作") {
		cat, urgency = "合作", "中"
	} else if strings.Contains(lower, "免费") || strings.Contains(lower, "中奖") {
		cat, urgency = "垃圾", "低"
	}

	return &runtime.ExecutionResult{
		Data: map[string]interface{}{
			"category": cat, "reason": fmt.Sprintf("基于关键词分析归类为%s", cat),
			"reply_suggestion": fmt.Sprintf("您好，已收到您的来信。关于「%s」的问题，我们会尽快处理。", truncate(input, 20)),
			"urgency": urgency,
		},
		TokenUsage: runtime.TokenUsage{InputTokens: 80, OutputTokens: 120},
	}, nil
}

func (p *EmailSorterPlugin) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	return &runtime.ReviewResult{
		Passed: true, Score: 5.0, ShouldRetry: false, Summary: "分类完成",
	}, nil
}

// XHSPosterPlugin - embedded Xiaohongshu content agent
type XHSPosterPlugin struct{}

func (p *XHSPosterPlugin) Name() string { return "xhs_poster" }

func (p *XHSPosterPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name: "xhs_poster", Summary: "生成小红书种草笔记内容",
		Version: "1.0.0", Tags: []string{"social-media", "xiaohongshu", "content-marketing"},
		RequiresHITL: true,
	}
}

func (p *XHSPosterPlugin) Execute(_ context.Context, input string, _ map[string]interface{}) (*runtime.ExecutionResult, error) {
	return &runtime.ExecutionResult{
		Data: map[string]interface{}{
			"title":            fmt.Sprintf("被问爆了！%s 真的绝了✨", truncate(input, 10)),
			"body":             fmt.Sprintf("姐妹们！今天一定要给你们安利这个%s\n\n真的被惊艳到了😍 品质超级棒，细节处理也很到位。\n\n推荐指数：⭐⭐⭐⭐⭐\n\n我已经回购好几次了，每次都被朋友夸！", input),
			"hashtags":         []string{"#好物推荐", "#种草", "#测评", "#生活好物", "#值得入手"},
			"image_suggestions": []string{"产品整体展示图", "细节特写图", "使用场景图"},
			"style":            "好物推荐",
		},
		TokenUsage: runtime.TokenUsage{InputTokens: 90, OutputTokens: 200},
	}, nil
}

func (p *XHSPosterPlugin) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	return &runtime.ReviewResult{
		Passed: true, Score: 5.0, ShouldRetry: false, Summary: "内容已生成",
	}, nil
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

func scoreFromChecks(checks []runtime.CheckResult) float32 {
	if len(checks) == 0 {
		return 5.0
	}
	passed := 0
	for _, c := range checks {
		if c.Passed {
			passed++
		}
	}
	return float32(passed) / float32(len(checks)) * 5.0
}
