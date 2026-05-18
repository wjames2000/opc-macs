package xiaohongshu

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
	baseURL    = "https://edith.xiaohongshu.com"
	apiVersion = "/api/v1"
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
			AuthURL:      "https://www.xiaohongshu.com/oauth/authorize",
			TokenURL:     "https://www.xiaohongshu.com/oauth/token",
			RedirectURI:  redirectURI,
			Scopes:       []string{"user_info", "note_publish"},
		}),
	}
}

func (c *Client) Name() string                { return "xiaohongshu" }
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
	return c.oauth.ValidateToken(ctx, token, "https://www.xiaohongshu.com/oauth/userinfo")
}

func (c *Client) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("xiaohongshu: auth token required")
	}

	body := map[string]interface{}{
		"title":       req.Content.Title,
		"desc":        req.Content.Body,
		"contentType": string(req.Content.ContentType),
		"tags":        req.Content.Tags,
	}
	if len(req.MediaFiles) > 0 {
		mediaURLs := make([]string, 0, len(req.MediaFiles))
		for _, m := range req.MediaFiles {
			if m.URL != "" {
				mediaURLs = append(mediaURLs, m.URL)
			} else {
				mediaURLs = append(mediaURLs, m.FilePath)
			}
		}
		body["mediaUrls"] = mediaURLs
	}
	data, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/note/publish", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("xiaohongshu: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.AuthToken.AccessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("xiaohongshu: publish: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			NoteID string `json:"note_id"`
			URL    string `json:"url"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("xiaohongshu: decode response: %w", err)
	}
	if !result.Success {
		return nil, fmt.Errorf("xiaohongshu: publish failed: %s", result.Message)
	}

	return &platform.PublishResponse{
		PlatformPostID: result.Data.NoteID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
		PlatformURL:    result.Data.URL,
		PostURL:        result.Data.URL,
	}, nil
}

func (c *Client) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/user/me", nil)
	if err != nil {
		return nil, fmt.Errorf("xiaohongshu: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("xiaohongshu: get account: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			UserID      string `json:"user_id"`
			Nickname    string `json:"nickname"`
			Avatar      string `json:"avatar"`
			Description string `json:"desc"`
			Followers   int64  `json:"followers"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("xiaohongshu: decode account: %w", err)
	}

	return &platform.Account{
		ID:             result.Data.UserID,
		PlatformName:   "xiaohongshu",
		PlatformUserID: result.Data.UserID,
		Name:           result.Data.Nickname,
		Avatar:         result.Data.Avatar,
		Bio:            result.Data.Description,
		FollowerCount:  result.Data.Followers,
	}, nil
}

func (c *Client) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/note/status?note_id="+platformPostID, nil)
	if err != nil {
		return nil, fmt.Errorf("xiaohongshu: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("xiaohongshu: get status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("xiaohongshu: read body: %w", err)
	}

	var status struct {
		Status string `json:"status"`
	}
	json.Unmarshal(body, &status)

	pubStatus := platform.PublishStatusPending
	switch status.Status {
	case "published":
		pubStatus = platform.PublishStatusPublished
	case "failed":
		pubStatus = platform.PublishStatusFailed
	}

	return &platform.PublishStatusInfo{
		PlatformPostID: platformPostID,
		Status:         pubStatus,
	}, nil
}

func (c *Client) ListAccounts(ctx context.Context) ([]*platform.Account, error) {
	acct, err := c.GetAccountInfo(ctx)
	if err != nil {
		return nil, err
	}
	return []*platform.Account{acct}, nil
}

func (c *Client) Capabilities() platform.Capabilities {
	return platform.Capabilities{
		SupportedContentTypes: []platform.ContentType{platform.ContentTypeImage, platform.ContentTypeVideo, platform.ContentTypeMixed},
		MaxImageCount:         18,
		MaxImageSize:          20 * 1024 * 1024,
		MaxVideoDuration:      300,
		MaxTextLength:         2000,
		MaxTitleLength:        20,
		MaxTagsCount:          10,
		SupportsSchedule:      false,
		SupportsMultiAccount:  true,
		SupportedAspectRatios: []string{"3:4", "1:1"},
	}
}
