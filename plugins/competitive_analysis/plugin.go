package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type CompetitiveAnalysisPlugin struct{}

func (p *CompetitiveAnalysisPlugin) Name() string { return "competitive_analysis" }

func (p *CompetitiveAnalysisPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:    "competitive_analysis",
		Summary: "对竞品进行 SWOT 结构化分析",
		Version: "1.0.0",
		Tags:    []string{"analysis", "strategy", "business", "SWOT"},
	}
}

var competitiveSystemPrompt = `你是资深的商业分析师，擅长竞品分析。请对用户提供的竞品信息进行结构化分析。

[输出格式 JSON]
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

func (p *CompetitiveAnalysisPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("competitive_analysis: 未配置模型 API")
	}
	modelName := resolveModelName(opts)

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: competitiveSystemPrompt,
		UserMessage:  input,
	})
	if err != nil {
		return nil, fmt.Errorf("competitive_analysis: 模型调用失败：%w", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		return nil, fmt.Errorf("competitive_analysis: 输出解析失败：%w", err)
	}

	return &runtime.ExecutionResult{
		Data:       output,
		TokenUsage: runtime.TokenUsage{InputTokens: resp.InputTokens, OutputTokens: resp.OutputTokens, ModelName: modelName},
		RawTrace:   resp.RawResponse,
	}, nil
}

func (p *CompetitiveAnalysisPlugin) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	return &runtime.ReviewResult{Passed: true, Score: 5.0, Summary: "competitive_analysis: 完成"}, nil
}

func resolveModelName(opts map[string]interface{}) string {
	if m, ok := opts["model_name"].(string); ok && m != "" {
		return m
	}
	return "deepseek-v4-flash"
}

var Agent CompetitiveAnalysisPlugin
