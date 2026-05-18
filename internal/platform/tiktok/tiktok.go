package tiktok

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
	baseURL    = "https://open.tiktokapis.com"
	apiVersion = "/v2"
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
			AuthURL:      "https://www.tiktok.com/v2/auth/authorize",
			TokenURL:     "https://open.tiktokapis.com/v2/oauth/token",
			RedirectURI:  redirectURI,
			Scopes:       []string{"user.info.basic", "video.upload", "video.publish"},
		}),
	}
}

func (c *Client) Name() string                { return "tiktok" }
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
	return c.oauth.ValidateToken(ctx, token, "https://open.tiktokapis.com/v2/user/info/")
}

func (c *Client) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("tiktok: auth token required")
	}

	form := url.Values{}
	form.Set("access_token", req.AuthToken.AccessToken)
	form.Set("title", req.Content.Title)
	form.Set("description", req.Content.Body)
	form.Set("privacy_level", "PUBLIC")

	if len(req.MediaFiles) > 0 {
		form.Set("media_url", req.MediaFiles[0].URL)
	}
	if len(req.Content.Tags) > 0 {
		tags, _ := json.Marshal(req.Content.Tags)
		form.Set("tags", string(tags))
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/video/publish/", bytes.NewReader([]byte(form.Encode())))
	if err != nil {
		return nil, fmt.Errorf("tiktok: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tiktok: publish: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tiktok: read response: %w", err)
	}

	var result struct {
		Data struct {
			PublishID string `json:"publish_id"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("tiktok: decode response: %w", err)
	}
	if result.Error.Code != "" {
		return nil, fmt.Errorf("tiktok: publish failed (%s): %s", result.Error.Code, result.Error.Message)
	}

	return &platform.PublishResponse{
		PlatformPostID: result.Data.PublishID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
	}, nil
}

func (c *Client) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/user/info/?fields=open_id,union_id,avatar_url,display_name,bio_description,follower_count", nil)
	if err != nil {
		return nil, fmt.Errorf("tiktok: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tiktok: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("tiktok: read body: %w", err)
	}

	var result struct {
		Data struct {
			User struct {
				OpenID      string `json:"open_id"`
				UnionID     string `json:"union_id"`
				DisplayName string `json:"display_name"`
				AvatarURL   string `json:"avatar_url"`
				BioDesc     string `json:"bio_description"`
				FollowerCnt int64  `json:"follower_count"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("tiktok: decode account: %w", err)
	}

	return &platform.Account{
		ID:             result.Data.User.OpenID,
		PlatformName:   "tiktok",
		PlatformUserID: result.Data.User.UnionID,
		Name:           result.Data.User.DisplayName,
		Avatar:         result.Data.User.AvatarURL,
		Bio:            result.Data.User.BioDesc,
		FollowerCount:  result.Data.User.FollowerCnt,
	}, nil
}

func (c *Client) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/video/publish/status/?publish_id="+platformPostID, nil)
	if err != nil {
		return nil, fmt.Errorf("tiktok: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tiktok: get status: %w", err)
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
		MaxVideoSize:          500 * 1024 * 1024,
		MaxTextLength:         4000,
		MaxTitleLength:        50,
		MaxTagsCount:          10,
		SupportsSchedule:      true,
		SupportsMultiAccount:  true,
		SupportedAspectRatios: []string{"9:16", "1:1"},
	}
}
