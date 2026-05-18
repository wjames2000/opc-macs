package pinterest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/wjames2000/opc-macs/internal/platform"
)

const (
	baseURL    = "https://api.pinterest.com"
	apiVersion = "/v5"
)

type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	oauth        *platform.OAuth2Client
}

func NewClient(clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		oauth: platform.NewOAuth2Client(platform.OAuth2Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			AuthURL:      "https://www.pinterest.com/oauth/",
			TokenURL:     "https://api.pinterest.com/v5/oauth/token",
			RedirectURI:  redirectURI,
			Scopes:       []string{"boards:read", "pins:read", "pins:write"},
		}),
	}
}

func (c *Client) Name() string                { return "pinterest" }
func (c *Client) AuthType() platform.AuthType { return platform.AuthTypeOAuth2 }

func (c *Client) GetAuthURL(redirectURI, state string) (string, error) {
	return c.oauth.GetAuthURL(state), nil
}

func (c *Client) ExchangeCode(ctx context.Context, code, redirectURI string) (*platform.AuthToken, error) {
	return c.oauth.ExchangeCode(ctx, code)
}

func (c *Client) RefreshToken(ctx context.Context, token *platform.AuthToken) (*platform.AuthToken, error) {
	return c.oauth.RefreshToken(ctx, token)
}

func (c *Client) ValidateToken(ctx context.Context, token *platform.AuthToken) (bool, error) {
	return c.oauth.ValidateToken(ctx, token, "https://api.pinterest.com/v5/user_account")
}

func (c *Client) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("pinterest: auth token required")
	}

	payload := map[string]interface{}{
		"title":       req.Content.Title,
		"description": req.Content.Body,
		"alt_text":    req.Content.Title,
	}

	if len(req.MediaFiles) > 0 {
		mf := req.MediaFiles[0]
		if mf.URL != "" {
			payload["image_url"] = mf.URL
		}
		if mf.FilePath != "" {
			payload["media_source"] = map[string]interface{}{
				"source_type": "image_url",
				"url":         mf.FilePath,
			}
		}
	}

	if len(req.Content.Tags) > 0 {
		payload["note"] = req.Content.Body + "\n\n" + "#" + req.Content.Tags[0]
	}

	if !req.ScheduleAt.IsZero() {
		payload["scheduled_for"] = req.ScheduleAt.Format(time.RFC3339)
	}

	data, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/pins", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("pinterest: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.AuthToken.AccessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("pinterest: publish: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("pinterest: read body: %w", err)
	}

	var result struct {
		Data struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("pinterest: decode response: %w", err)
	}
	if result.Data.ID == "" {
		return nil, fmt.Errorf("pinterest: publish failed: %s", result.Message)
	}

	return &platform.PublishResponse{
		PlatformPostID: result.Data.ID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
		PlatformURL:    result.Data.URL,
	}, nil
}

func (c *Client) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/user_account", nil)
	if err != nil {
		return nil, fmt.Errorf("pinterest: create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.clientID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("pinterest: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("pinterest: read body: %w", err)
	}

	var result struct {
		Data struct {
			ID        string `json:"id"`
			Username  string `json:"username"`
			FullName  string `json:"full_name"`
			ImageURL  string `json:"profile_image"`
			Bio       string `json:"biography"`
			Followers int64  `json:"follower_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("pinterest: decode account: %w", err)
	}

	return &platform.Account{
		ID:             result.Data.ID,
		PlatformName:   "pinterest",
		PlatformUserID: result.Data.ID,
		Name:           result.Data.FullName,
		Avatar:         result.Data.ImageURL,
		Bio:            result.Data.Bio,
		FollowerCount:  result.Data.Followers,
	}, nil
}

func (c *Client) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/pins/"+platformPostID, nil)
	if err != nil {
		return nil, fmt.Errorf("pinterest: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("pinterest: get status: %w", err)
	}
	defer resp.Body.Close()

	return &platform.PublishStatusInfo{
		PlatformPostID: platformPostID,
		Status:         platform.PublishStatusPublished,
	}, nil
}

func (c *Client) ListAccounts(ctx context.Context) ([]*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/user_account", nil)
	if err != nil {
		return nil, fmt.Errorf("pinterest: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("pinterest: list accounts: %w", err)
	}
	defer resp.Body.Close()

	return nil, fmt.Errorf("pinterest: single user account only")
}

func (c *Client) Capabilities() platform.Capabilities {
	return platform.Capabilities{
		SupportedContentTypes: []platform.ContentType{platform.ContentTypeImage},
		MaxImageCount:         1,
		MaxImageSize:          20 * 1024 * 1024,
		MaxTextLength:         500,
		MaxTitleLength:        100,
		MaxTagsCount:          20,
		SupportsSchedule:      true,
		SupportsMultiAccount:  true,
		SupportedAspectRatios: []string{"2:3", "1:1"},
	}
}
