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

// KlingClient implements video generation via the Kling AI API.
// Supports text-to-video and image-to-video generation.
type KlingClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewKlingClient creates a new Kling AI API client.
func NewKlingClient(apiKey string) *KlingClient {
	return &KlingClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: "https://api.klingai.com/v1",
	}
}

// GenerateVideo submits a text-to-video generation job.
// Returns the job ID for status polling.
func (c *KlingClient) GenerateVideo(ctx context.Context, prompt string, opts map[string]interface{}) (string, error) {
	body := map[string]interface{}{
		"model":            "kling-v1",
		"prompt":           prompt,
		"mode":             "text2video",
		"duration_seconds": 5,
	}
	for k, v := range opts {
		body[k] = v
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("kling marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/video/generate", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("kling create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("kling request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("kling read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("kling API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			JobID string `json:"job_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("kling parse response: %w", err)
	}

	if result.Data.JobID == "" {
		return "", fmt.Errorf("kling: empty job_id in response")
	}

	return result.Data.JobID, nil
}

// GenerateImageToVideo submits an image-to-video generation job.
// imageURL is the URL of the source image, prompt describes the motion.
func (c *KlingClient) GenerateImageToVideo(ctx context.Context, imageURL, prompt string, opts map[string]interface{}) (string, error) {
	body := map[string]interface{}{
		"model":            "kling-v1",
		"image_url":        imageURL,
		"prompt":           prompt,
		"mode":             "img2video",
		"duration_seconds": 5,
	}
	for k, v := range opts {
		body[k] = v
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("kling img2video marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/video/generate", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("kling img2video request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("kling img2video failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("kling img2video read: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("kling img2video error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			JobID string `json:"job_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("kling img2video parse: %w", err)
	}

	if result.Data.JobID == "" {
		return "", fmt.Errorf("kling img2video: empty job_id")
	}

	return result.Data.JobID, nil
}

// CheckStatus queries the current status of a Kling video generation job.
func (c *KlingClient) CheckStatus(ctx context.Context, jobID string) (*GenerationResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/video/status/"+jobID, nil)
	if err != nil {
		return nil, fmt.Errorf("kling status request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kling status failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kling read status: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kling status error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Data struct {
			Status string `json:"status"`
			URL    string `json:"video_url"`
			Error  string `json:"error,omitempty"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("kling parse status: %w", err)
	}

	status := StatusProcessing
	switch apiResp.Data.Status {
	case "completed", "succeeded":
		status = StatusCompleted
	case "failed":
		status = StatusFailed
	case "pending":
		status = StatusPending
	}

	return &GenerationResult{
		ID:        jobID,
		MediaType: MediaTypeVideo,
		Status:    status,
		URL:       apiResp.Data.URL,
		Error:     apiResp.Data.Error,
		CreatedAt: time.Now(),
	}, nil
}
