package agent

import (
	"context"
	"fmt"
)

func callLLM(ctx context.Context, model, prompt string) (string, error) {
	// MVP: direct HTTP call to Gemini API
	// In production, use ADK Go or LiteLLM
	return "", fmt.Errorf("LLM call not implemented in MVP stub - model=%s", model)
}
