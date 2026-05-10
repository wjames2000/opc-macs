package main

import (
	"context"
	"fmt"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

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

func (p *CopywriterPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	return &runtime.ExecutionResult{
		Data: map[string]interface{}{
			"short_copy":  "智能水杯，懂你温度",
			"long_copy":   "这是一款革命性的智能水杯...（LLM 生成内容，此处为 MVP stub）",
			"social_copy": "有了它，喝水变成一种享受 💧",
			"style":       "科技简约",
		},
		TokenUsage: runtime.TokenUsage{InputTokens: 0, OutputTokens: 0},
	}, nil
}

func (p *CopywriterPlugin) Review(ctx context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output format"}, nil
	}

	checks := []runtime.CheckResult{}

	if short, ok := data["short_copy"].(string); ok {
		checks = append(checks, runtime.CheckResult{
			Item: "short_copy length ≤ 20", Passed: len(short) <= 20, Detail: fmt.Sprintf("len=%d", len(short)),
		})
	}
	if long, ok := data["long_copy"].(string); ok {
		checks = append(checks, runtime.CheckResult{
			Item: "long_copy length 100-200", Passed: len(long) >= 100 && len(long) <= 200,
			Detail: fmt.Sprintf("len=%d", len(long)),
		})
	}
	if social, ok := data["social_copy"].(string); ok {
		checks = append(checks, runtime.CheckResult{
			Item: "social_copy length ≤ 80", Passed: len(social) <= 80, Detail: fmt.Sprintf("len=%d", len(social)),
		})
	}

	allPassed := true
	for _, c := range checks {
		if !c.Passed {
			allPassed = false
			break
		}
	}

	return &runtime.ReviewResult{
		Passed:       allPassed,
		Score:        calculateScore(checks),
		CheckResults: checks,
		Summary:      summaryFromChecks(checks),
		ShouldRetry:  !allPassed,
	}, nil
}

var Agent CopywriterPlugin

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

func summaryFromChecks(checks []runtime.CheckResult) string {
	passed, total := 0, len(checks)
	for _, c := range checks {
		if c.Passed {
			passed++
		}
	}
	return fmt.Sprintf("%d/%d checks passed", passed, total)
}
