package wechat

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
	baseURL = "https://api.weixin.qq.com"
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
			AuthURL:      "https://open.weixin.qq.com/connect/qrconnect",
			TokenURL:     "https://api.weixin.qq.com/sns/oauth2/access_token",
			RedirectURI:  redirectURI,
			Scopes:       []string{"snsapi_userinfo"},
		}),
	}
}

func (c *Client) Name() string                { return "wechat" }
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
	return c.oauth.ValidateToken(ctx, token, "https://api.weixin.qq.com/sns/userinfo")
}

func (c *Client) doRequest(ctx context.Context, method, urlPath string, body []byte) (*http.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, method, baseURL+urlPath, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(httpReq)
}

func (c *Client) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("wechat: auth token required")
	}

	accessToken := req.AuthToken.AccessToken
	switch req.Content.ContentType {
	case platform.ContentTypeArticle:
		return c.publishArticle(ctx, accessToken, req)
	case platform.ContentTypeVideo:
		return c.publishVideo(ctx, accessToken, req)
	default:
		return c.publishArticle(ctx, accessToken, req)
	}
}

func (c *Client) publishArticle(ctx context.Context, accessToken string, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	payload := map[string]interface{}{
		"title":                 req.Content.Title,
		"content":               req.Content.Body,
		"need_open_comment":     1,
		"only_fans_can_comment": 0,
	}
	body, _ := json.Marshal(payload)
	path := "/cgi-bin/draft/add?access_token=" + accessToken

	resp, err := c.doRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, fmt.Errorf("wechat: publish article: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
		MediaID string `json:"media_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("wechat: decode response: %w", err)
	}
	if result.Errcode != 0 {
		return nil, fmt.Errorf("wechat: publish failed (%d): %s", result.Errcode, result.Errmsg)
	}

	return &platform.PublishResponse{
		PlatformPostID: result.MediaID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
		PlatformURL:    fmt.Sprintf("https://mp.weixin.qq.com/s?media_id=%s", result.MediaID),
	}, nil
}

func (c *Client) publishVideo(ctx context.Context, accessToken string, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	path := "/cgi-bin/material/add_material?access_token=" + accessToken + "&type=video"
	var fileData []byte
	if len(req.MediaFiles) > 0 && req.MediaFiles[0].FilePath != "" {
		fileData = []byte(req.MediaFiles[0].FilePath)
	}

	resp, err := c.doRequest(ctx, "POST", path, fileData)
	if err != nil {
		return nil, fmt.Errorf("wechat: upload video: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
		MediaID string `json:"media_id"`
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("wechat: read body: %w", err)
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("wechat: decode upload: %w", err)
	}
	if result.Errcode != 0 {
		return nil, fmt.Errorf("wechat: upload failed (%d): %s", result.Errcode, result.Errmsg)
	}

	return &platform.PublishResponse{
		PlatformPostID: result.MediaID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
	}, nil
}

func (c *Client) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	resp, err := c.doRequest(ctx, "GET", "/sns/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("wechat: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("wechat: read body: %w", err)
	}

	var result struct {
		OpenID   string `json:"openid"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"headimgurl"`
		UnionID  string `json:"unionid"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("wechat: decode account: %w", err)
	}

	return &platform.Account{
		ID:             result.OpenID,
		PlatformName:   "wechat",
		PlatformUserID: result.OpenID,
		Name:           result.Nickname,
		Avatar:         result.Avatar,
		PlatformExtra: map[string]interface{}{
			"unionid": result.UnionID,
		},
	}, nil
}

func (c *Client) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
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
		SupportedContentTypes: []platform.ContentType{platform.ContentTypeVideo, platform.ContentTypeArticle},
		MaxVideoDuration:      600,
		MaxVideoSize:          2 * 1024 * 1024 * 1024,
		MaxTextLength:         50000,
		MaxTitleLength:        64,
		MaxTagsCount:          0,
		SupportsSchedule:      true,
		SupportsMultiAccount:  true,
		SupportedAspectRatios: []string{"16:9", "9:16"},
	}
}
