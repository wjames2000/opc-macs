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
	modelName := "gemini-2.0-flash"

	if client == nil {
		// Dev mode fallback
		return devCopywrite(input), nil
	}

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

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: copywriterSystemPrompt,
		UserMessage:  sb.String(),
	})
	if err != nil {
		return devCopywrite(input), nil
	}

	// Try to parse as JSON; fallback to structured text
	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		output = devCopywrite(input).Data.(map[string]interface{})
	}

	return &runtime.ExecutionResult{
		Data: output,
		TokenUsage: runtime.TokenUsage{
			InputTokens:  resp.InputTokens,
			OutputTokens: resp.OutputTokens,
			ModelName:    modelName,
		},
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

func devCopywrite(input string) *runtime.ExecutionResult {
	// Intelligent dev mode: generate context-aware output from keywords in input
	product := extractProductName(input)
	style := detectStyle(input)

	return &runtime.ExecutionResult{
		Data: map[string]interface{}{
			"short_copy":  fmt.Sprintf("终于等到%s了！🔥 %s全攻略", product, product),
			"long_copy":   fmt.Sprintf("【%s 深度测评】\n\n最近被问爆的%s，今天来给大家详细说说。\n\n首先说外观：设计真的很有质感，拿在手里就知道了。\n\n再说功能：该有的全都有，细节处理很到位。\n\n最后说性价比：这个价位段里，绝对是值得入手的选择。\n\n总之，如果你正在考虑%s，这篇文章应该能帮你做决定。", product, product, product),
			"social_copy": fmt.Sprintf("终于入手了%s！之前观望了好久，用了一周只想说——真香！😍 #好物分享 #值得入手", product),
			"style":       style,
		},
		TokenUsage: runtime.TokenUsage{ModelName: "dev-mode"},
	}
}

// extractProductName tries to extract the product/brand name from input
func extractProductName(input string) string {
	// Remove common prefixes
	cleaned := input
	prefixes := []string{"帮我写", "帮我", "写一个", "写一篇", "关于", "的推广文案", "的文案", "的笔记", "的广告", "的营销文章", "推荐", "推广"}
	for _, p := range prefixes {
		cleaned = strings.ReplaceAll(cleaned, p, "")
	}
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return "这款产品"
	}
	// Take the first meaningful part
	words := strings.Fields(cleaned)
	if len(words) > 3 {
		cleaned = strings.Join(words[:3], " ")
	}
	return cleaned
}

// detectStyle returns a style based on input keywords
func detectStyle(input string) string {
	styles := []struct {
		keywords []string
		style    string
	}{
		{[]string{"科技", "数码", "智能", "手机", "电脑", "AI", "app", "软件"}, "科技简约"},
		{[]string{"美食", "食品", "零食", "餐厅", "菜", "吃", "喝"}, "温暖亲和"},
		{[]string{"教育", "课程", "培训", "学习", "书"}, "专业正式"},
		{[]string{"时尚", "穿搭", "美妆", "护肤", "衣服", "包"}, "年轻活力"},
	}
	lower := strings.ToLower(input)
	for _, s := range styles {
		for _, kw := range s.keywords {
			if strings.Contains(lower, strings.ToLower(kw)) {
				return s.style
			}
		}
	}
	return "温暖亲和"
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
	modelName := "gemini-2.0-flash"

	// Always use keyword pre-check
	preCategory := quickSpamCheck(input)
	if preCategory != "" {
		return &runtime.ExecutionResult{
			Data: map[string]interface{}{
				"category":         preCategory,
				"reason":           "命中垃圾关键词规则",
				"reply_suggestion": "",
				"urgency":          "低",
			},
			TokenUsage: runtime.TokenUsage{ModelName: "rule-based"},
		}, nil
	}

	if client == nil {
		return devClassify(input), nil
	}

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: emailSorterSystemPrompt,
		UserMessage:  input,
	})
	if err != nil {
		return devClassify(input), nil
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		output = devClassify(input).Data.(map[string]interface{})
	}

	return &runtime.ExecutionResult{
		Data: output,
		TokenUsage: runtime.TokenUsage{
			InputTokens:  resp.InputTokens,
			OutputTokens: resp.OutputTokens,
			ModelName:    modelName,
		},
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

func quickSpamCheck(body string) string {
	lower := strings.ToLower(body)
	spamWords := []string{"免费", "中奖", "转账", "点击链接", "free", "winner"}
	for _, w := range spamWords {
		if strings.Contains(lower, w) {
			return "垃圾"
		}
	}
	return ""
}

func devClassify(input string) *runtime.ExecutionResult {
	lower := strings.ToLower(input)
	cat, urgency, reason := "咨询", "低", "常规咨询"

	if strings.Contains(lower, "投诉") || strings.Contains(lower, "退款") || strings.Contains(lower, "赔偿") {
		cat, urgency, reason = "投诉", "高", "检测到投诉/退款关键词"
	} else if strings.Contains(lower, "合作") || strings.Contains(lower, "商务") || strings.Contains(lower, "partner") {
		cat, urgency, reason = "合作", "中", "检测到合作意向"
	} else if strings.Contains(lower, "免费") || strings.Contains(lower, "中奖") || strings.Contains(lower, "转账") {
		cat, urgency, reason = "垃圾", "低", "命中垃圾邮件规则"
	} else if strings.Contains(lower, "咨询") || strings.Contains(lower, "请问") || strings.Contains(lower, "help") || strings.Contains(lower, "how") {
		cat, urgency, reason = "咨询", "中", "检测到咨询内容"
	}

	reply := ""
	switch cat {
	case "投诉":
		reply = fmt.Sprintf("尊敬的客户，\n\n非常抱歉给您带来不便。我们已经收到您的投诉（「%s」），正在加急处理中，预计24小时内会有专人联系您。\n\n感谢您的耐心与理解。", extractProductName(input))
	case "合作":
		reply = fmt.Sprintf("您好，\n\n感谢您的合作意向！我们非常期待与您进一步沟通。\n\n关于%s，请提供以下信息以便我们更好地了解您的需求：\n1. 公司/个人简介\n2. 合作方式设想\n3. 联系方式", extractProductName(input))
	case "咨询":
		reply = fmt.Sprintf("您好，\n\n感谢您的来信。关于您咨询的问题，我们的回复如下：\n\n「%s」\n\n如有其他疑问，欢迎随时联系我们。", input)
	case "垃圾":
		reply = ""
	}

	return &runtime.ExecutionResult{
		Data: map[string]interface{}{
			"category":         cat,
			"reason":           reason,
			"reply_suggestion": reply,
			"urgency":          urgency,
		},
		TokenUsage: runtime.TokenUsage{ModelName: "dev-mode"},
	}
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
	modelName := "gemini-2.0-flash"

	if client == nil {
		return devXHSPost(input), nil
	}

	memories, _ := opts["memories"].([]string)
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

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: xhsSystemPrompt,
		UserMessage:  sb.String(),
	})
	if err != nil {
		return devXHSPost(input), nil
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		output = devXHSPost(input).Data.(map[string]interface{})
	}

	return &runtime.ExecutionResult{
		Data: output,
		TokenUsage: runtime.TokenUsage{
			InputTokens:  resp.InputTokens,
			OutputTokens: resp.OutputTokens,
			ModelName:    modelName,
		},
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

func devXHSPost(input string) *runtime.ExecutionResult {
	product := extractProductName(input)
	return &runtime.ExecutionResult{
		Data: map[string]interface{}{
			"title":            fmt.Sprintf("姐妹们！%s真的太香了💕", truncate(product, 12)),
			"body":             fmt.Sprintf("姐妹们！今天来给大家安利一下%s！\n\n先说结论：真的值得入！✨\n\n🌟 颜值：包装设计就很高级，拿在手里质感满满\n🌟 使用感：第一次用就被惊艳到了，细节做得很好\n🌟 性价比：在同价位里绝对是天花板级别的\n\n有条件的姐妹一定要试试！保证不后悔！💯\n\n#好物分享 #真实测评 #值得入手", product),
			"hashtags":         []string{"#好物推荐", "#种草", "#真实测评", "#值得入手", "#我的好物清单"},
			"image_suggestions": []string{"产品整体展示图", "使用效果实拍图", "产品细节特写"},
			"style":            "好物推荐",
		},
		TokenUsage: runtime.TokenUsage{ModelName: "dev-mode"},
	}
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

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
