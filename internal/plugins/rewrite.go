package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

var rewriteSystemPrompt = `你是专业的内容改写专家。你的任务是根据目标风格改写内容。

[角色边界]
- 只做内容改写，不生成新内容
- 保持原文核心信息和事实不变
- 不添加原文不存在的观点或数据

[风格选项]
- professional: 专业正式风格，适合商务场景
- casual: 轻松口语风格，适合社交平台
- concise: 精炼风格，删除冗余保留核心
- detailed: 详细扩展风格，补充背景信息
- persuasive: 说服力风格，适合营销推广
- academic: 学术风格，严谨引用

[输出要求]
- original: 原文摘要（前100字）
- style: 使用风格
- rewritten: 改写后的完整内容
- changes: 主要改动说明列表
- preserved: 保留的核心信息列表`

type RewritePlugin struct{}

func (p *RewritePlugin) Name() string { return "rewrite" }

func (p *RewritePlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "rewrite",
		Summary:      "多风格内容改写工具，支持专业/口语/精炼/详细/说服力/学术六种风格",
		Version:      "1.0.0",
		Tags:         []string{"rewrite", "content", "style-transfer"},
		RequiresHITL: true,
		ModelName:    "gemini-2.0-flash",
	}
}

func (p *RewritePlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("rewrite: 未配置模型 API")
	}
	modelName := resolveModelName(opts, p.Info().ModelName)

	var req struct {
		Content        string   `json:"content"`
		Style          string   `json:"style"`
		TargetAudience string   `json:"target_audience,omitempty"`
		Keywords       []string `json:"keywords,omitempty"`
	}
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		return nil, fmt.Errorf("rewrite: invalid input: %w", err)
	}
	if req.Content == "" {
		return nil, fmt.Errorf("rewrite: content is empty")
	}
	validStyles := map[string]bool{"professional": true, "casual": true, "concise": true, "detailed": true, "persuasive": true, "academic": true}
	if req.Style == "" || !validStyles[req.Style] {
		req.Style = "professional"
	}

	var sb strings.Builder
	sb.WriteString(rewriteSystemPrompt)
	sb.WriteString("\n\n目标风格: ")
	sb.WriteString(req.Style)
	if req.TargetAudience != "" {
		sb.WriteString("\n目标受众: ")
		sb.WriteString(req.TargetAudience)
	}
	if len(req.Keywords) > 0 {
		sb.WriteString("\n关键词: ")
		sb.WriteString(strings.Join(req.Keywords, ", "))
	}
	sb.WriteString("\n\n原文内容:\n")
	if len(req.Content) > 2000 {
		sb.WriteString(req.Content[:2000])
		sb.WriteString("\n...（原文过长，已截取前2000字）")
	} else {
		sb.WriteString(req.Content)
	}

	modelResp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: "你是一个专业内容改写专家。严格按照 system prompt 的格式输出 JSON。",
		UserMessage:  sb.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("rewrite: model call: %w", err)
	}

	result := map[string]interface{}{
		"style":    req.Style,
		"audience": req.TargetAudience,
		"keywords": req.Keywords,
		"output":   modelResp.Content,
	}

	return &runtime.ExecutionResult{
		Data: result,
		TokenUsage: runtime.TokenUsage{
			InputTokens:  modelResp.InputTokens,
			OutputTokens: modelResp.OutputTokens,
			ModelName:    modelName,
		},
	}, nil
}

func (p *RewritePlugin) Review(ctx context.Context, output interface{}) (*runtime.ReviewResult, error) {
	data, ok := output.(map[string]interface{})
	if !ok {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "output type mismatch"}, nil
	}
	style, _ := data["style"].(string)
	outputText, _ := data["output"].(string)
	if style == "" || outputText == "" {
		return &runtime.ReviewResult{Passed: false, Score: 0, Summary: "missing style or output"}, nil
	}
	return &runtime.ReviewResult{
		Passed:  true,
		Score:   0.9,
		Summary: fmt.Sprintf("rewritten in %s style, %d chars", style, len(outputText)),
	}, nil
}
