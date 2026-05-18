package facebook

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
	baseURL    = "https://graph.facebook.com"
	apiVersion = "/v19.0"
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
			AuthURL:      "https://www.facebook.com/dialog/oauth",
			TokenURL:     "https://graph.facebook.com/v19.0/oauth/access_token",
			RedirectURI:  redirectURI,
			Scopes:       []string{"pages_manage_posts", "pages_read_engagement", "instagram_basic", "instagram_content_publish"},
		}),
	}
}

func (c *Client) Name() string                { return "facebook" }
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
	return c.oauth.ValidateToken(ctx, token, "https://graph.facebook.com/v19.0/me")
}

func (c *Client) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("facebook: auth token required")
	}

	form := url.Values{}
	form.Set("access_token", req.AuthToken.AccessToken)
	form.Set("message", req.Content.Body)

	if len(req.MediaFiles) > 0 {
		mf := req.MediaFiles[0]
		if mf.URL != "" {
			form.Set("url", mf.URL)
		}
		switch mf.MediaType {
		case "video":
			return c.publishVideo(ctx, form)
		default:
			form.Set("caption", req.Content.Title)
			return c.publishPhoto(ctx, form)
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/me/feed", bytes.NewReader([]byte(form.Encode())))
	if err != nil {
		return nil, fmt.Errorf("facebook: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("facebook: publish: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("facebook: read body: %w", err)
	}

	var result struct {
		ID    string `json:"id"`
		Error struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("facebook: decode response: %w", err)
	}
	if result.Error.Code != 0 {
		return nil, fmt.Errorf("facebook: publish failed: %s", result.Error.Message)
	}

	return &platform.PublishResponse{
		PlatformPostID: result.ID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
		PlatformURL:    fmt.Sprintf("https://www.facebook.com/%s", result.ID),
	}, nil
}

func (c *Client) publishPhoto(ctx context.Context, form url.Values) (*platform.PublishResponse, error) {
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/me/photos", bytes.NewReader([]byte(form.Encode())))
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("facebook: publish photo: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		ID    string `json:"id"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	json.Unmarshal(body, &result)

	return &platform.PublishResponse{
		PlatformPostID: result.ID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
	}, nil
}

func (c *Client) publishVideo(ctx context.Context, form url.Values) (*platform.PublishResponse, error) {
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/me/videos", bytes.NewReader([]byte(form.Encode())))
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("facebook: publish video: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		ID string `json:"id"`
	}
	json.Unmarshal(body, &result)

	return &platform.PublishResponse{
		PlatformPostID: result.ID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
	}, nil
}

func (c *Client) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/me?fields=id,name,picture,about,followers_count", nil)
	if err != nil {
		return nil, fmt.Errorf("facebook: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("facebook: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("facebook: read body: %w", err)
	}

	var result struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		About   string `json:"about"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
		FollowersCount int64 `json:"followers_count"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("facebook: decode account: %w", err)
	}

	return &platform.Account{
		ID:             result.ID,
		PlatformName:   "facebook",
		PlatformUserID: result.ID,
		Name:           result.Name,
		Avatar:         result.Picture.Data.URL,
		Bio:            result.About,
		FollowerCount:  result.FollowersCount,
	}, nil
}

func (c *Client) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/"+platformPostID+"?fields=status_type,created_time", nil)
	if err != nil {
		return nil, fmt.Errorf("facebook: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("facebook: get status: %w", err)
	}
	defer resp.Body.Close()

	return &platform.PublishStatusInfo{
		PlatformPostID: platformPostID,
		Status:         platform.PublishStatusPublished,
	}, nil
}

func (c *Client) ListAccounts(ctx context.Context) ([]*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/me/accounts?fields=id,name,picture,access_token", nil)
	if err != nil {
		return nil, fmt.Errorf("facebook: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("facebook: list accounts: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("facebook: read body: %w", err)
	}

	var result struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("facebook: decode accounts: %w", err)
	}

	accounts := make([]*platform.Account, 0, len(result.Data))
	for _, d := range result.Data {
		accounts = append(accounts, &platform.Account{
			ID:             d.ID,
			PlatformName:   "facebook",
			PlatformUserID: d.ID,
			Name:           d.Name,
		})
	}
	return accounts, nil
}

func (c *Client) Capabilities() platform.Capabilities {
	return platform.Capabilities{
		SupportedContentTypes: []platform.ContentType{platform.ContentTypeImage, platform.ContentTypeVideo, platform.ContentTypeText},
		MaxImageCount:         10,
		MaxVideoDuration:      14400,
		MaxVideoSize:          10 * 1024 * 1024 * 1024,
		MaxTextLength:         63206,
		MaxTitleLength:        80,
		MaxTagsCount:          5,
		SupportsSchedule:      true,
		SupportsMultiAccount:  true,
	}
}
