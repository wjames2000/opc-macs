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

func NewModelClient(provider, apiBaseURL, apiKey string) (runtime.ModelClient, error) {
	baseURL := apiBaseURL
	if baseURL == "" {
		switch provider {
		case "openai":
			baseURL = "https://api.openai.com"
		case "deepseek":
			baseURL = "https://api.deepseek.com"
		case "qwen":
			baseURL = "https://dashscope.aliyuncs.com/compatible-mode"
		case "kimi":
			baseURL = "https://api.moonshot.cn"
		case "claude":
			baseURL = "https://api.anthropic.com"
		case "gemini":
			baseURL = "https://generativelanguage.googleapis.com"
		default:
			return nil, fmt.Errorf("未知 provider '%s'，请在 config.yaml 中设置 model.api_base_url", provider)
		}
	}
	return &httpClient{
		baseURL:  baseURL,
		apiKey:   apiKey,
		provider: provider,
		client:   &http.Client{Timeout: 60 * time.Second},
	}, nil
}

type httpClient struct {
	baseURL  string
	apiKey   string
	provider string
	client   *http.Client
}

type openAIMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIReq struct {
	Model       string      `json:"model"`
	Messages    []openAIMsg `json:"messages"`
	Temperature float32     `json:"temperature"`
	MaxTokens   int         `json:"max_tokens"`
}

type openAIResp struct {
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

type claudeMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeReq struct {
	Model       string      `json:"model"`
	System      string      `json:"system,omitempty"`
	MaxTokens   int         `json:"max_tokens"`
	Temperature float32     `json:"temperature"`
	Messages    []claudeMsg `json:"messages"`
}

type claudeResp struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (c *httpClient) Call(ctx context.Context, req runtime.ModelRequest) (*runtime.ModelResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("API Key 未配置：请在 config.yaml 中设置 model.api_key")
	}
	if req.Model == "" {
		req.Model = "deepseek-chat"
	}
	if c.provider == "claude" {
		return c.callClaude(ctx, req)
	}
	return c.callOpenAI(ctx, req)
}

func (c *httpClient) callOpenAI(ctx context.Context, req runtime.ModelRequest) (*runtime.ModelResponse, error) {
	chatReq := openAIReq{
		Model: req.Model,
		Messages: []openAIMsg{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserMessage},
		},
		Temperature: 0.3,
		MaxTokens:   4096,
	}
	body, _ := json.Marshal(chatReq)

	url := c.baseURL + "/v1/chat/completions"
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("[%s] 连接失败：%w", c.provider, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[%s] HTTP %d：%s", c.provider, resp.StatusCode, string(respBody))
	}

	var chatResp openAIResp
	json.Unmarshal(respBody, &chatResp)
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("[%s] 返回空结果", c.provider)
	}

	return &runtime.ModelResponse{
		Content:      chatResp.Choices[0].Message.Content,
		InputTokens:  chatResp.Usage.PromptTokens,
		OutputTokens: chatResp.Usage.CompletionTokens,
		RawResponse:  string(respBody),
	}, nil
}

func (c *httpClient) callClaude(ctx context.Context, req runtime.ModelRequest) (*runtime.ModelResponse, error) {
	claudeReqData := claudeReq{
		Model:       req.Model,
		System:      req.SystemPrompt,
		MaxTokens:   4096,
		Temperature: 0.3,
		Messages:    []claudeMsg{{Role: "user", Content: req.UserMessage}},
	}
	body, _ := json.Marshal(claudeReqData)

	url := c.baseURL + "/v1/messages"
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("[claude] 连接失败：%w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[claude] HTTP %d：%s", resp.StatusCode, string(respBody))
	}

	var claudeRespData claudeResp
	json.Unmarshal(respBody, &claudeRespData)
	if len(claudeRespData.Content) == 0 {
		return nil, fmt.Errorf("[claude] 返回空内容")
	}

	return &runtime.ModelResponse{
		Content:      claudeRespData.Content[0].Text,
		InputTokens:  claudeRespData.Usage.InputTokens,
		OutputTokens: claudeRespData.Usage.OutputTokens,
		RawResponse:  string(respBody),
	}, nil
}

type openAIEmbedReq struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type openAIEmbedResp struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
	} `json:"usage"`
}

func (c *httpClient) Embed(ctx context.Context, req runtime.EmbedRequest) (*runtime.EmbedResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("API Key 未配置：请在 config.yaml 中设置 model.api_key")
	}
	embedReq := openAIEmbedReq{Model: req.Model, Input: req.Input}
	if embedReq.Model == "" {
		embedReq.Model = "text-embedding-3-small"
	}
	body, _ := json.Marshal(embedReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/embeddings", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("[%s] embedding 连接失败：%w", c.provider, err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("[%s] embedding HTTP %d：%s", c.provider, resp.StatusCode, string(respBody))
	}
	var embedResp openAIEmbedResp
	json.Unmarshal(respBody, &embedResp)
	if len(embedResp.Data) == 0 {
		return nil, fmt.Errorf("[%s] embedding 返回空结果", c.provider)
	}
	return &runtime.EmbedResponse{
		Embedding:   embedResp.Data[0].Embedding,
		InputTokens: embedResp.Usage.PromptTokens,
	}, nil
}
