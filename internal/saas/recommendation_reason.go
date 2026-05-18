package saas

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

// RecommendationReasoner generates LLM-powered explanations for task recommendations.
// It wraps the plugin system to produce rich, contextual reasons for why a specific
// marketplace task is recommended to a creator.
type RecommendationReasoner struct {
	loader *runtime.Loader
}

// NewRecommendationReasoner creates a new reasoner that uses the given plugin loader.
func NewRecommendationReasoner(loader *runtime.Loader) *RecommendationReasoner {
	return &RecommendationReasoner{loader: loader}
}

// GenerateReasons enriches each recommendation's Reason field with LLM-generated text.
// Falls back to the original reason if the plugin is unavailable or the LLM call fails.
func (r *RecommendationReasoner) GenerateReasons(ctx context.Context, recommendations []TaskRecommendation) ([]TaskRecommendation, error) {
	if r.loader == nil {
		return recommendations, nil
	}

	plugin, ok := r.loader.Get("content_creator")
	if !ok {
		return recommendations, nil
	}

	results := make([]TaskRecommendation, len(recommendations))
	copy(results, recommendations)

	for i, rec := range results {
		if rec.Task == nil {
			continue
		}

		prompt := r.buildReasonPrompt(rec.Task)

		input, err := json.Marshal(map[string]interface{}{
			"action":     "generate_text",
			"prompt":     prompt,
			"max_length": 150,
		})
		if err != nil {
			continue
		}

		result, err := plugin.Execute(ctx, string(input), map[string]interface{}{})
		if err != nil {
			continue
		}

		dataMap, ok := result.Data.(map[string]interface{})
		if !ok {
			continue
		}
		reasonText, ok := dataMap["text"].(string)
		if !ok || reasonText == "" {
			continue
		}

		results[i].Reason = r.parseReasonResponse(reasonText)
	}

	return results, nil
}

// GenerateBatchReasons processes recommendations in batches of the given size.
// This prevents excessive LLM calls for large recommendation lists.
func (r *RecommendationReasoner) GenerateBatchReasons(ctx context.Context, recs []TaskRecommendation, batchSize int) ([]TaskRecommendation, error) {
	if batchSize <= 0 {
		batchSize = 5
	}

	results := make([]TaskRecommendation, len(recs))
	copy(results, recs)

	for i := 0; i < len(results); i += batchSize {
		end := i + batchSize
		if end > len(results) {
			end = len(results)
		}

		batch, err := r.GenerateReasons(ctx, results[i:end])
		if err != nil {
			// continue with what we have
			continue
		}
		copy(results[i:end], batch)
	}

	return results, nil
}

// buildReasonPrompt constructs a structured prompt for the LLM to generate a recommendation reason.
func (r *RecommendationReasoner) buildReasonPrompt(task *MarketplaceTask) string {
	var sb strings.Builder

	sb.WriteString("You are a marketplace recommendation assistant. ")
	sb.WriteString("Explain in ONE short sentence (Chinese, max 50 characters) why this task suits a creator.")

	sb.WriteString("\n\nTask details:\n")
	sb.WriteString(fmt.Sprintf("- Title: %s\n", task.Title))
	if task.Description != "" {
		desc := task.Description
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		sb.WriteString(fmt.Sprintf("- Description: %s\n", desc))
	}
	sb.WriteString(fmt.Sprintf("- Budget: ¥%.0f\n", task.Budget))
	if task.Platform != "" {
		sb.WriteString(fmt.Sprintf("- Platform: %s\n", task.Platform))
	}
	if !task.Deadline.IsZero() {
		sb.WriteString(fmt.Sprintf("- Deadline: %s\n", task.Deadline.Format("2006-01-02")))
	}

	sb.WriteString("\nResponse format: Just the reason sentence, nothing else.")
	return sb.String()
}

// parseReasonResponse extracts clean reason text from the LLM response.
// Strips any markdown formatting, quotes, or extra whitespace.
func (r *RecommendationReasoner) parseReasonResponse(response string) string {
	text := strings.TrimSpace(response)
	text = strings.TrimPrefix(text, "\"")
	text = strings.TrimSuffix(text, "\"")
	text = strings.TrimPrefix(text, "'")
	text = strings.TrimSuffix(text, "'")
	text = strings.TrimSpace(text)
	return text
}
