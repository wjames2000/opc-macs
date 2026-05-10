package main

import (
	"context"
	"fmt"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type XHSPosterPlugin struct{}

func (p *XHSPosterPlugin) Name() string { return "xhs_poster" }

func (p *XHSPosterPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "xhs_poster",
		Summary:      "生成小红书种草笔记内容",
		Version:      "1.0.0",
		Tags:         []string{"social-media", "xiaohongshu", "content-marketing"},
		RequiresHITL: true,
	}
}

func (p *XHSPosterPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	return &runtime.ExecutionResult{
		Data: map[string]interface{}{
			"title":            "这款好物真的绝了✨",
			"body":             "亲测好用！分享一下最近入手的这款产品...（LLM 生成内容，此处为 MVP stub）\n\n真的被惊艳到了！😍 从包装到使用体验都非常棒。\n\n推荐指数：⭐⭐⭐⭐⭐",
			"hashtags":         []string{"#好物推荐", "#种草", "#测评", "#生活好物", "#值得入手"},
			"image_suggestions": []string{"产品全家福展示图", "使用效果对比图", "产品细节特写"},
			"style":            "好物推荐",
		},
		TokenUsage: runtime.TokenUsage{InputTokens: 0, OutputTokens: 0},
	}, nil
}

func (p *XHSPosterPlugin) Review(ctx context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output format"}, nil
	}

	checks := []runtime.CheckResult{}
	allPassed := true

	if title, ok := data["title"].(string); ok {
		passed := len(title) <= 20 && containsEmoji(title)
		if !passed {
			allPassed = false
		}
		checks = append(checks, runtime.CheckResult{
			Item: "title ≤ 20 chars with emoji", Passed: passed,
			Detail: fmt.Sprintf("len=%d, hasEmoji=%v", len(title), containsEmoji(title)),
		})
	}

	if body, ok := data["body"].(string); ok {
		passed := len(body) >= 200 && len(body) <= 500
		if !passed {
			allPassed = false
		}
		checks = append(checks, runtime.CheckResult{
			Item: "body 200-500 chars", Passed: passed,
			Detail: fmt.Sprintf("len=%d", len(body)),
		})
	}

	if tags, ok := data["hashtags"].([]interface{}); ok {
		passed := len(tags) >= 5 && len(tags) <= 10
		if !passed {
			allPassed = false
		}
		checks = append(checks, runtime.CheckResult{
			Item: "hashtags 5-10", Passed: passed,
			Detail: fmt.Sprintf("count=%d", len(tags)),
		})
	}

	return &runtime.ReviewResult{
		Passed:       allPassed,
		Score:        calculateScore(checks),
		CheckResults: checks,
		Summary:      summaryStr(checks),
		ShouldRetry:  !allPassed,
	}, nil
}

func containsEmoji(s string) bool {
	for _, r := range s {
		if r > 0x1F300 && r < 0x1FA00 {
			return true
		}
	}
	return false
}

func calculateScore(checks []runtime.CheckResult) float32 {
	if len(checks) == 0 {
		return 0
	}
	passed := 0
	for _, c := range checks {
		if c.Passed {
			passed++
		}
	}
	return float32(passed) / float32(len(checks)) * 5.0
}

func summaryStr(checks []runtime.CheckResult) string {
	passed, total := 0, len(checks)
	for _, c := range checks {
		if c.Passed {
			passed++
		}
	}
	return fmt.Sprintf("%d/%d checks passed", passed, total)
}

var Agent XHSPosterPlugin
