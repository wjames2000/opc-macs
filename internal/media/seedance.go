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

// SeedanceClient implements video generation via the Seedance API (text-to-video).
type SeedanceClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewSeedanceClient creates a new Seedance API client.
func NewSeedanceClient(apiKey string) *SeedanceClient {
	return &SeedanceClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: "https://api.seedance.ai/v1",
	}
}

// GenerateVideo submits a text-to-video generation job.
// Returns the job ID for status polling.
func (c *SeedanceClient) GenerateVideo(ctx context.Context, prompt string, opts map[string]interface{}) (string, error) {
	body := map[string]interface{}{
		"prompt": prompt,
	}
	for k, v := range opts {
		body[k] = v
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("seedance marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/video/generate", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("seedance create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("seedance request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("seedance read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("seedance API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			JobID string `json:"job_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("seedance parse response: %w", err)
	}

	if result.Data.JobID == "" {
		return "", fmt.Errorf("seedance: empty job_id in response")
	}

	return result.Data.JobID, nil
}

// CheckStatus queries the current status of a video generation job.
func (c *SeedanceClient) CheckStatus(ctx context.Context, jobID string) (*GenerationResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/video/status/"+jobID, nil)
	if err != nil {
		return nil, fmt.Errorf("seedance status request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("seedance status failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("seedance read status: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("seedance status error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Data struct {
			Status string `json:"status"`
			URL    string `json:"video_url"`
			Error  string `json:"error,omitempty"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("seedance parse status: %w", err)
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
