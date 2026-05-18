package media

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DallEClient implements the ImageGenerator interface via the OpenAI DALL-E 3 API.
type DallEClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewDallEClient creates a new DALL-E 3 API client.
func NewDallEClient(apiKey string) *DallEClient {
	return &DallEClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		baseURL: "https://api.openai.com/v1",
	}
}

// Generate creates an image from a text prompt using DALL-E 3.
// opts can include: "size" (string, e.g. "1024x1024", "1792x1024"),
// "quality" (string, "standard" or "hd"), "style" (string, "vivid" or "natural").
// Returns the image bytes (PNG format).
func (c *DallEClient) Generate(ctx context.Context, prompt string, opts map[string]interface{}) ([]byte, error) {
	body := map[string]interface{}{
		"model":           "dall-e-3",
		"prompt":          prompt,
		"n":               1,
		"size":            "1024x1024",
		"quality":         "standard",
		"response_format": "b64_json",
	}
	if opts != nil {
		if size, ok := opts["size"]; ok {
			body["size"] = size
		}
		if quality, ok := opts["quality"]; ok {
			body["quality"] = quality
		}
		if style, ok := opts["style"]; ok {
			body["style"] = style
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("dalle marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/images/generations", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("dalle create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dalle request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("dalle read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dalle API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			B64JSON       string `json:"b64_json"`
			URL           string `json:"url"`
			RevisedPrompt string `json:"revised_prompt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("dalle parse: %w", err)
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("dalle: no images in response")
	}

	if result.Data[0].B64JSON != "" {
		decoded := make([]byte, len(result.Data[0].B64JSON)*3/4)
		n, err := decodeBase64(result.Data[0].B64JSON, decoded)
		if err != nil {
			return nil, fmt.Errorf("dalle decode base64: %w", err)
		}
		return decoded[:n], nil
	}

	if result.Data[0].URL != "" {
		imgResp, err := http.Get(result.Data[0].URL)
		if err != nil {
			return nil, fmt.Errorf("dalle fetch url: %w", err)
		}
		defer imgResp.Body.Close()
		return io.ReadAll(imgResp.Body)
	}

	return nil, fmt.Errorf("dalle: no image data in response")
}

// decodeBase64 decodes a base64-encoded string into the provided buffer.
func decodeBase64(s string, buf []byte) (int, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	pos := 0
	val := 0
	bits := -6
	for _, r := range s {
		if r == '=' {
			break
		}
		idx := -1
		for i, c := range alphabet {
			if c == r {
				idx = i
				break
			}
		}
		if idx < 0 {
			continue
		}
		val = (val << 6) | idx
		bits += 6
		if bits >= 0 {
			if pos < len(buf) {
				buf[pos] = byte((val >> uint(bits)) & 0xFF)
				pos++
			}
			bits -= 8
			val &= (1 << (bits + 8)) - 1
		}
	}
	return pos, nil
}
