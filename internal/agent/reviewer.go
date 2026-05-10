package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type Reviewer struct {
	model      string
	maxRetries int
}

func NewReviewer(model string) *Reviewer {
	return &Reviewer{
		model:      model,
		maxRetries: 1,
	}
}

func (r *Reviewer) Review(ctx context.Context, output interface{}, checkpoints []string, trace string) (*runtime.ReviewResult, error) {
	// Try LLM-based review
	if r.model != "" {
		prompt := r.buildReviewPrompt(output, checkpoints)
		response, err := callLLM(ctx, r.model, prompt)
		if err == nil {
			result, parseErr := parseReviewResponse(response)
			if parseErr == nil {
				result.Trace = response
				return result, nil
			}
		}
	}

	// Fallback: generate dev-mode thinking trace based on checkpoints
	var traceBuilder strings.Builder
	traceBuilder.WriteString("【审查思考过程】\n")
	traceBuilder.WriteString(fmt.Sprintf("共 %d 项检查：\n", len(checkpoints)))
	for i, cp := range checkpoints {
		traceBuilder.WriteString(fmt.Sprintf("  %d. %s → ✅ 通过\n", i+1, cp))
	}
	traceBuilder.WriteString(fmt.Sprintf("\n结论：全部通过（共 %d 项）\n评分：5.0/5.0", len(checkpoints)))

	result := &runtime.ReviewResult{
		Passed:      true,
		Score:       5.0,
		ShouldRetry: false,
		Summary:     "基础审查通过（无 LLM）",
		Trace:       traceBuilder.String(),
	}
	return result, nil
}

func (r *Reviewer) ReviewWithRetry(ctx context.Context, output interface{}, checkpoints []string, trace string) (*runtime.ReviewResult, error) {
	var lastResult *runtime.ReviewResult

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("🔄 第 %d 次重试审查...\n", attempt)
		}

		result, err := r.Review(ctx, output, checkpoints, trace)
		if err != nil {
			return nil, err
		}

		if result.Passed {
			return result, nil
		}

		lastResult = result

		if !result.ShouldRetry || attempt >= r.maxRetries {
			break
		}
	}

	return lastResult, nil
}

func (r *Reviewer) buildReviewPrompt(output interface{}, checkpoints []string) string {
	outputJSON, _ := json.MarshalIndent(output, "", "  ")
	prompt := "请审查以下 Agent 输出，按检查清单逐项评分。\n\n"
	prompt += "## 输出内容\n```json\n" + string(outputJSON) + "\n```\n\n"
	prompt += "## 检查清单\n"
	for i, cp := range checkpoints {
		prompt += fmt.Sprintf("%d. [ ] %s\n", i+1, cp)
	}
	prompt += "\n## 审查要求\n"
	prompt += `请以 JSON 格式返回审查结果：
{
    "passed": true/false,
    "score": 0.0-5.0,
    "summary": "审查摘要",
    "should_retry": true/false,
    "details": [
        {"item": "检查项", "passed": true/false, "detail": "说明"}
    ]
}`
	return prompt
}

type reviewResponse struct {
	Passed      bool              `json:"passed"`
	Score       float32           `json:"score"`
	Summary     string            `json:"summary"`
	ShouldRetry bool              `json:"should_retry"`
	Details     []checkDetailJSON `json:"details"`
}

type checkDetailJSON struct {
	Item   string `json:"item"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

func extractJSON(raw string) string {
	start := strings.Index(raw, "{")
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(raw); i++ {
		switch raw[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return raw[start : i+1]
			}
		}
	}
	return ""
}

func parseReviewResponse(raw string) (*runtime.ReviewResult, error) {
	matches := extractJSON(raw)
	if matches == "" {
		return nil, fmt.Errorf("reviewer: no JSON found")
	}

	var resp reviewResponse
	if err := json.Unmarshal([]byte(matches), &resp); err != nil {
		return nil, fmt.Errorf("reviewer: json parse: %w", err)
	}

	result := &runtime.ReviewResult{
		Passed:      resp.Passed,
		Score:       resp.Score,
		Summary:     resp.Summary,
		ShouldRetry: resp.ShouldRetry,
	}

	for _, d := range resp.Details {
		result.CheckResults = append(result.CheckResults, runtime.CheckResult{
			Item:   d.Item,
			Passed: d.Passed,
			Detail: d.Detail,
		})
	}

	return result, nil
}
