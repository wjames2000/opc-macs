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

// MidjourneyClient implements the ImageGenerator interface via the Midjourney API.
// Supports imagine, vary, and upscale operations for AI image generation.
type MidjourneyClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewMidjourneyClient creates a new Midjourney API client.
func NewMidjourneyClient(apiKey string) *MidjourneyClient {
	return &MidjourneyClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		baseURL: "https://api.midjourney.com/v1",
	}
}

// Generate creates an image from a text prompt using Midjourney's imagine command.
// opts can include: "aspect_ratio" (string, e.g. "16:9"), "style" (string),
// "stylize" (int, 0-1000), "chaos" (int, 0-100), "version" (string, e.g. "6").
// Returns the image bytes (PNG format).
func (c *MidjourneyClient) Generate(ctx context.Context, prompt string, opts map[string]interface{}) ([]byte, error) {
	body := map[string]interface{}{
		"action": "imagine",
		"prompt": prompt,
	}
	if opts != nil {
		for k, v := range opts {
			body[k] = v
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("midjourney marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/image/generate", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("midjourney create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("midjourney request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("midjourney read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("midjourney API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			ImageData []byte `json:"image_data"`
			ImageURL  string `json:"image_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("midjourney parse: %w", err)
	}

	if len(result.Data.ImageData) > 0 {
		return result.Data.ImageData, nil
	}

	if result.Data.ImageURL != "" {
		imgResp, err := http.Get(result.Data.ImageURL)
		if err != nil {
			return nil, fmt.Errorf("midjourney fetch image: %w", err)
		}
		defer imgResp.Body.Close()
		return io.ReadAll(imgResp.Body)
	}

	return nil, fmt.Errorf("midjourney: no image data or url in response")
}

// Vary creates a variation of an existing image.
// imageID is the ID returned from a previous Generate call.
func (c *MidjourneyClient) Vary(ctx context.Context, imageID string, opts map[string]interface{}) ([]byte, error) {
	body := map[string]interface{}{
		"action":   "vary",
		"image_id": imageID,
	}
	for k, v := range opts {
		body[k] = v
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("midjourney vary marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/image/vary", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("midjourney vary request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("midjourney vary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("midjourney vary error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}

// Upscale increases the resolution of a generated image.
// imageID is the ID from Generate, scale is the multiplier (e.g. 2, 4).
func (c *MidjourneyClient) Upscale(ctx context.Context, imageID string, scale int) ([]byte, error) {
	body := map[string]interface{}{
		"action":   "upscale",
		"image_id": imageID,
		"scale":    scale,
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("midjourney upscale marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/image/upscale", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("midjourney upscale request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("midjourney upscale: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("midjourney upscale error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}
