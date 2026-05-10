package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

// NewDefaultModelClient creates a model client for the configured API endpoint.
// If baseURL is empty, returns a no-op client (development mode).
func NewDefaultModelClient(baseURL, apiKey string) runtime.ModelClient {
	if baseURL == "" {
		return &noopClient{}
	}
	return &httpClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type noopClient struct{}

func (c *noopClient) Call(ctx context.Context, req runtime.ModelRequest) (*runtime.ModelResponse, error) {
	return nil, fmt.Errorf("dev mode: no API configured, set model.api_base_url in config.yaml")
}

type httpClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float32       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (c *httpClient) Call(ctx context.Context, req runtime.ModelRequest) (*runtime.ModelResponse, error) {
	chatReq := chatRequest{
		Model: req.Model,
		Messages: []chatMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserMessage},
		},
		Temperature: 0.3,
		MaxTokens:   4096,
	}

	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("model: marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("model: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("model: http call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("model: read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("model: API error %d: %s", resp.StatusCode, string(respBody))
	}

	// Save raw response for thinking trace
	rawStr := string(respBody)

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("model: parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("model: no choices in response")
	}

	return &runtime.ModelResponse{
		Content:      chatResp.Choices[0].Message.Content,
		InputTokens:  chatResp.Usage.PromptTokens,
		OutputTokens: chatResp.Usage.CompletionTokens,
		RawResponse:  rawStr,
	}, nil
}

func truncateStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
