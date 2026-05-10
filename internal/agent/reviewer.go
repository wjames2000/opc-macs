package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

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

func (r *Reviewer) Review(ctx context.Context, output interface{}, checkpoints []string) (*runtime.ReviewResult, error) {
	prompt := r.buildReviewPrompt(output, checkpoints)
	response, err := callLLM(ctx, r.model, prompt)
	if err != nil {
		return &runtime.ReviewResult{
			Passed:      false,
			Score:       0,
			ShouldRetry: true,
			Summary:     "审查过程异常",
		}, nil
	}

	result, err := parseReviewResponse(response)
	if err != nil {
		return &runtime.ReviewResult{
			Passed:      false,
			Score:       0,
			ShouldRetry: true,
			Summary:     "审查结果解析失败",
		}, nil
	}

	return result, nil
}

func (r *Reviewer) ReviewWithRetry(ctx context.Context, output interface{}, checkpoints []string) (*runtime.ReviewResult, error) {
	var lastResult *runtime.ReviewResult

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			fmt.Printf("🔄 第 %d 次重试审查...\n", attempt)
		}

		result, err := r.Review(ctx, output, checkpoints)
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

var reviewJSONRe = regexp.MustCompile(`\{[^{}]*\}`)

func parseReviewResponse(raw string) (*runtime.ReviewResult, error) {
	matches := reviewJSONRe.FindString(raw)
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
