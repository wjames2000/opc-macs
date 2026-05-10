package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

var spamKeywords = []string{
	"免费", "中奖", "转账", "点击链接",
	"free", "winner", "bank transfer",
}

type EmailSorterPlugin struct{}

func (p *EmailSorterPlugin) Name() string { return "email_sorter" }

func (p *EmailSorterPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "email_sorter",
		Summary:      "分析邮件内容并分类，生成回复建议",
		Version:      "1.0.0",
		Tags:         []string{"email", "customer-service", "classification"},
		RequiresHITL: true,
	}
}

func (p *EmailSorterPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	category, reason, urgency := classifyEmail(input)
	reply := generateReply(category, input)

	return &runtime.ExecutionResult{
		Data: map[string]interface{}{
			"category":         category,
			"reason":           reason,
			"reply_suggestion": reply,
			"urgency":          urgency,
		},
		TokenUsage: runtime.TokenUsage{InputTokens: 0, OutputTokens: 0},
	}, nil
}

func classifyEmail(body string) (category, reason, urgency string) {
	lower := strings.ToLower(body)

	if quickSpamCheck(body) {
		return "垃圾", "命中垃圾关键词", "低"
	}

	if strings.Contains(lower, "投诉") || strings.Contains(lower, "退款") ||
		strings.Contains(lower, "complaint") || strings.Contains(lower, "refund") {
		return "投诉", "检测到投诉/退款关键词", "高"
	}
	if strings.Contains(lower, "合作") || strings.Contains(lower, "partner") ||
		strings.Contains(lower, "合作") {
		return "合作", "检测到合作意向", "中"
	}
	if strings.Contains(lower, "咨询") || strings.Contains(lower, "请问") ||
		strings.Contains(lower, "question") || strings.Contains(lower, "how") {
		return "咨询", "检测到咨询类内容", "中"
	}
	return "咨询", "默认分类", "低"
}

func generateReply(category, input string) string {
	switch category {
	case "投诉":
		return fmt.Sprintf("您好，非常抱歉给您带来不便。我们已收到您的反馈，正在加急处理。请提供更多信息以便我们尽快解决您的问题。")
	case "合作":
		return fmt.Sprintf("您好，感谢您的合作意向！我们会尽快安排专人对接沟通具体合作细节。")
	case "咨询":
		return fmt.Sprintf("您好，感谢您的咨询！关于您提出的问题，我们的回复如下：\n\n（请根据具体问题补充回复内容）\n\n如有其他疑问，欢迎随时联系我们。")
	case "垃圾":
		return ""
	default:
		return "您好，已收到您的邮件，我们会尽快处理。"
	}
}

func quickSpamCheck(body string) bool {
	bodyLower := strings.ToLower(body)
	for _, kw := range spamKeywords {
		if strings.Contains(bodyLower, kw) {
			return true
		}
	}
	return false
}

func (p *EmailSorterPlugin) Review(ctx context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "invalid output"}, nil
	}

	cat, _ := data["category"].(string)
	validCategories := map[string]bool{"咨询": true, "投诉": true, "合作": true, "垃圾": true}

	return &runtime.ReviewResult{
		Passed:      validCategories[cat],
		Score:       5.0,
		CheckResults: []runtime.CheckResult{{Item: "category valid", Passed: validCategories[cat]}},
		Summary:     fmt.Sprintf("category=%s", cat),
		ShouldRetry: !validCategories[cat],
	}, nil
}

var Agent EmailSorterPlugin
