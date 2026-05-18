package twitter

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
	baseURL    = "https://api.twitter.com"
	apiVersion = "/2"
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
			AuthURL:      "https://twitter.com/i/oauth2/authorize",
			TokenURL:     "https://api.twitter.com/2/oauth2/token",
			RedirectURI:  redirectURI,
			Scopes:       []string{"tweet.read", "tweet.write", "users.read", "offline.access"},
		}),
	}
}

func (c *Client) Name() string                { return "twitter" }
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
	return c.oauth.ValidateToken(ctx, token, "https://api.twitter.com/2/users/me")
}

func (c *Client) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("twitter: auth token required")
	}

	payload := map[string]interface{}{
		"text": req.Content.Body,
	}

	if req.Content.Title != "" {
		payload["text"] = req.Content.Title + "\n\n" + req.Content.Body
	}

	if len(req.MediaFiles) > 0 {
		mediaIDs := make([]string, 0, len(req.MediaFiles))
		for _, mf := range req.MediaFiles {
			mid, err := c.uploadMedia(ctx, req.AuthToken.AccessToken, &mf)
			if err != nil {
				return nil, fmt.Errorf("twitter: upload media: %w", err)
			}
			mediaIDs = append(mediaIDs, mid)
		}
		payload["media"] = map[string]interface{}{
			"media_ids": mediaIDs,
		}
	}

	data, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/tweets", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("twitter: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+req.AuthToken.AccessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("twitter: publish: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("twitter: read response: %w", err)
	}

	var result struct {
		Data struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("twitter: decode response: %w", err)
	}
	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("twitter: publish failed: %s", result.Errors[0].Message)
	}

	return &platform.PublishResponse{
		PlatformPostID: result.Data.ID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
		PlatformURL:    fmt.Sprintf("https://twitter.com/i/web/status/%s", result.Data.ID),
	}, nil
}

func (c *Client) uploadMedia(ctx context.Context, accessToken string, mf *platform.MediaFile) (string, error) {
	initPayload := map[string]interface{}{
		"total_bytes":    1024,
		"media_type":     mf.MimeType,
		"media_category": "tweet_image",
	}
	if mf.MediaType == "video" {
		initPayload["media_category"] = "tweet_video"
	}
	data, _ := json.Marshal(initPayload)

	initReq, _ := http.NewRequestWithContext(ctx, "POST", "https://upload.twitter.com/1.1/media/upload.json", bytes.NewReader(data))
	initReq.Header.Set("Content-Type", "application/json")
	initReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(initReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	return "media_" + mf.MediaType, nil
}

func (c *Client) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/users/me?user.fields=id,name,username,profile_image_url,description,public_metrics", nil)
	if err != nil {
		return nil, fmt.Errorf("twitter: create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.clientID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("twitter: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("twitter: read body: %w", err)
	}

	var result struct {
		Data struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Username    string `json:"username"`
			AvatarURL   string `json:"profile_image_url"`
			Description string `json:"description"`
			Metrics     struct {
				Followers int64 `json:"followers_count"`
			} `json:"public_metrics"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("twitter: decode account: %w", err)
	}

	return &platform.Account{
		ID:             result.Data.ID,
		PlatformName:   "twitter",
		PlatformUserID: result.Data.ID,
		Name:           result.Data.Name,
		Bio:            result.Data.Description,
		Avatar:         result.Data.AvatarURL,
		FollowerCount:  result.Data.Metrics.Followers,
		PlatformExtra: map[string]interface{}{
			"username": result.Data.Username,
		},
	}, nil
}

func (c *Client) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/tweets/"+platformPostID+"?tweet.fields=public_metrics", nil)
	if err != nil {
		return nil, fmt.Errorf("twitter: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("twitter: get status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("twitter: read body: %w", err)
	}

	var result struct {
		Data struct {
			ID      string `json:"id"`
			Metrics struct {
				RetweetCount int64 `json:"retweet_count"`
				LikeCount    int64 `json:"like_count"`
				ReplyCount   int64 `json:"reply_count"`
			} `json:"public_metrics"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("twitter: decode status: %w", err)
	}

	return &platform.PublishStatusInfo{
		PlatformPostID: result.Data.ID,
		Status:         platform.PublishStatusPublished,
		ViewCount:      result.Data.Metrics.RetweetCount,
		LikeCount:      result.Data.Metrics.LikeCount,
		CommentCount:   result.Data.Metrics.ReplyCount,
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
		SupportedContentTypes: []platform.ContentType{platform.ContentTypeText, platform.ContentTypeImage, platform.ContentTypeVideo},
		MaxImageCount:         4,
		MaxVideoDuration:      140,
		MaxVideoSize:          512 * 1024 * 1024,
		MaxTextLength:         280,
		MaxTagsCount:          0,
		SupportsSchedule:      true,
		SupportsMultiAccount:  true,
	}
}
