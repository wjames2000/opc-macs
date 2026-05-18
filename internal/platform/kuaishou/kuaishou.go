package kuaishou

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
	baseURL    = "https://open.kuaishou.com"
	apiVersion = "/rest"
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
			AuthURL:      "https://open.kuaishou.com/oauth/authorize",
			TokenURL:     "https://open.kuaishou.com/oauth/token",
			RedirectURI:  redirectURI,
			Scopes:       []string{"user_info", "video_publish"},
		}),
	}
}

func (c *Client) Name() string                { return "kuaishou" }
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
	return c.oauth.ValidateToken(ctx, token, "https://open.kuaishou.com/oauth/userinfo")
}

func (c *Client) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("kuaishou: auth token required")
	}

	form := url.Values{}
	form.Set("access_token", req.AuthToken.AccessToken)
	form.Set("title", req.Content.Title)
	form.Set("caption", req.Content.Body)
	if len(req.MediaFiles) > 0 {
		form.Set("media_url", req.MediaFiles[0].URL)
	}
	for i, tag := range req.Content.Tags {
		form.Set(fmt.Sprintf("tags[%d]", i), tag)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/video/publish", bytes.NewReader([]byte(form.Encode())))
	if err != nil {
		return nil, fmt.Errorf("kuaishou: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("kuaishou: publish: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kuaishou: read body: %w", err)
	}

	var result struct {
		Result   int    `json:"result"`
		ErrorMsg string `json:"error_msg"`
		Data     struct {
			PhotoID string `json:"photo_id"`
			URL     string `json:"url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("kuaishou: decode response: %w", err)
	}
	if result.Result != 1 {
		return nil, fmt.Errorf("kuaishou: publish failed: %s", result.ErrorMsg)
	}

	return &platform.PublishResponse{
		PlatformPostID: result.Data.PhotoID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
		PlatformURL:    result.Data.URL,
	}, nil
}

func (c *Client) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/openapi/user_info", nil)
	if err != nil {
		return nil, fmt.Errorf("kuaishou: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("kuaishou: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kuaishou: read body: %w", err)
	}

	var result struct {
		Result   int    `json:"result"`
		ErrorMsg string `json:"error_msg"`
		Data     struct {
			UserID   string `json:"user_id"`
			UserName string `json:"user_name"`
			Avatar   string `json:"avatar"`
			Bio      string `json:"bio"`
			Fans     int64  `json:"fans"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("kuaishou: decode account: %w", err)
	}
	if result.Result != 1 {
		return nil, fmt.Errorf("kuaishou: get account failed: %s", result.ErrorMsg)
	}

	return &platform.Account{
		ID:             result.Data.UserID,
		PlatformName:   "kuaishou",
		PlatformUserID: result.Data.UserID,
		Name:           result.Data.UserName,
		Avatar:         result.Data.Avatar,
		Bio:            result.Data.Bio,
		FollowerCount:  result.Data.Fans,
	}, nil
}

func (c *Client) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/photo/status?photo_id="+platformPostID, nil)
	if err != nil {
		return nil, fmt.Errorf("kuaishou: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("kuaishou: get status: %w", err)
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
		SupportedContentTypes: []platform.ContentType{platform.ContentTypeVideo},
		MaxVideoDuration:      600,
		MaxVideoSize:          4 * 1024 * 1024 * 1024,
		MaxTextLength:         500,
		MaxTitleLength:        25,
		MaxTagsCount:          8,
		SupportsSchedule:      false,
		SupportsMultiAccount:  true,
		SupportedAspectRatios: []string{"9:16", "16:9"},
	}
}
