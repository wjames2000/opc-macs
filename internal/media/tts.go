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

const (
	// VoiceFemale is a standard female TTS voice.
	VoiceFemale = "female"
	// VoiceMale is a standard male TTS voice.
	VoiceMale = "male"
)

// TTSClient generates speech from text via a text-to-speech API.
type TTSClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewTTSClient creates a new TTS API client.
func NewTTSClient(apiKey string) *TTSClient {
	return &TTSClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		baseURL: "https://api.tts-service.com/v1",
	}
}

// Synthesize converts text to speech audio.
// voice selects the voice (use VoiceFemale or VoiceMale constants).
// opts can include: "speed" (float64, 0.5-2.0), "pitch" (float64, -12 to 12).
// Returns the audio data as MP3 bytes.
func (c *TTSClient) Synthesize(ctx context.Context, text string, voice string, opts map[string]interface{}) ([]byte, error) {
	if text == "" {
		return nil, fmt.Errorf("tts: empty text")
	}
	if voice == "" {
		voice = VoiceFemale
	}

	body := map[string]interface{}{
		"text":   text,
		"voice":  voice,
		"format": "mp3",
	}
	for k, v := range opts {
		body[k] = v
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("tts marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/synthesize", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("tts create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tts request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tts API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tts read audio: %w", err)
	}

	return audioData, nil
}
