package bilibili

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/wjames2000/opc-macs/internal/platform"
)

const (
	baseURL    = "https://api.bilibili.com"
	apiVersion = "/x"
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
			AuthURL:      "https://api.bilibili.com/x/account/oauth2/authorize",
			TokenURL:     "https://api.bilibili.com/x/account/oauth2/token",
			RedirectURI:  redirectURI,
			Scopes:       []string{"user_info", "video_upload", "dynamic_send"},
		}),
	}
}

func (c *Client) Name() string                { return "bilibili" }
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
	return c.oauth.ValidateToken(ctx, token, "https://api.bilibili.com/x/space/myinfo")
}

func (c *Client) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("bilibili: auth token required")
	}

	form := url.Values{}
	form.Set("title", req.Content.Title)
	form.Set("content", req.Content.Body)
	form.Set("access_token", req.AuthToken.AccessToken)

	if len(req.MediaFiles) > 0 {
		form.Set("media_url", req.MediaFiles[0].URL)
	}
	for i, tag := range req.Content.Tags {
		form.Set(fmt.Sprintf("tag[%d]", i), tag)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/dynamic/create", bytes.NewReader([]byte(form.Encode())))
	if err != nil {
		return nil, fmt.Errorf("bilibili: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("bilibili: publish: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			DynamicID string `json:"dynamic_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("bilibili: decode response: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("bilibili: publish failed (%d): %s", result.Code, result.Message)
	}

	return &platform.PublishResponse{
		PlatformPostID: result.Data.DynamicID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
		PlatformURL:    fmt.Sprintf("https://t.bilibili.com/%s", result.Data.DynamicID),
	}, nil
}

func (c *Client) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/space/myinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("bilibili: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("bilibili: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("bilibili: read body: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Mid      int64  `json:"mid"`
			Name     string `json:"name"`
			Face     string `json:"face"`
			Sign     string `json:"sign"`
			Fans     int64  `json:"fans"`
			Follower int64  `json:"follower"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("bilibili: decode account: %w", err)
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("bilibili: get account failed (%d): %s", result.Code, result.Message)
	}

	followers := result.Data.Fans
	if followers == 0 {
		followers = result.Data.Follower
	}

	return &platform.Account{
		ID:             fmt.Sprintf("%d", result.Data.Mid),
		PlatformName:   "bilibili",
		PlatformUserID: fmt.Sprintf("%d", result.Data.Mid),
		Name:           result.Data.Name,
		Avatar:         result.Data.Face,
		Bio:            result.Data.Sign,
		FollowerCount:  followers,
	}, nil
}

func (c *Client) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/dynamic/detail?id="+platformPostID, nil)
	if err != nil {
		return nil, fmt.Errorf("bilibili: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("bilibili: get status: %w", err)
	}
	defer resp.Body.Close()

	return &platform.PublishStatusInfo{
		PlatformPostID: platformPostID,
		Status:         platform.PublishStatusPublished,
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
		SupportedContentTypes: []platform.ContentType{platform.ContentTypeVideo, platform.ContentTypeText},
		MaxVideoDuration:      600,
		MaxVideoSize:          8 * 1024 * 1024 * 1024,
		MaxTextLength:         3000,
		MaxTitleLength:        50,
		MaxTagsCount:          6,
		SupportsSchedule:      true,
		SupportsMultiAccount:  true,
		SupportedAspectRatios: []string{"16:9"},
	}
}
