package youtube

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
	baseURL    = "https://www.googleapis.com"
	apiVersion = "/youtube/v3"
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
			AuthURL:      "https://accounts.google.com/o/oauth2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			RedirectURI:  redirectURI,
			Scopes:       []string{"https://www.googleapis.com/auth/youtube.upload", "https://www.googleapis.com/auth/youtube.readonly"},
		}),
	}
}

func (c *Client) Name() string                { return "youtube" }
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
	return c.oauth.ValidateToken(ctx, token, "https://www.googleapis.com/oauth2/v1/tokeninfo")
}

func (c *Client) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("youtube: auth token required")
	}

	body := map[string]interface{}{
		"snippet": map[string]interface{}{
			"title":       req.Content.Title,
			"description": req.Content.Body,
			"tags":        req.Content.Tags,
		},
		"status": map[string]interface{}{
			"privacyStatus": "public",
		},
	}
	data, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		baseURL+"/upload"+apiVersion+"/videos?part=snippet,status&uploadType=resumable",
		bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("youtube: create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json; charset=UTF-8")
	httpReq.Header.Set("Authorization", "Bearer "+req.AuthToken.AccessToken)

	if len(req.MediaFiles) > 0 {
		httpReq.Header.Set("X-Upload-Content-Type", req.MediaFiles[0].MimeType)
		httpReq.Header.Set("X-Upload-Content-Length", fmt.Sprintf("%d", len(req.MediaFiles[0].FilePath)))
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("youtube: publish: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("youtube: read response: %w", err)
	}

	var videoResult struct {
		ID    string `json:"id"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(bodyBytes, &videoResult); err != nil {
		return nil, fmt.Errorf("youtube: decode response: %w", err)
	}
	if videoResult.Error.Code != 0 {
		return nil, fmt.Errorf("youtube: publish failed (%d): %s", videoResult.Error.Code, videoResult.Error.Message)
	}

	return &platform.PublishResponse{
		PlatformPostID: videoResult.ID,
		Status:         platform.PublishStatusPublished,
		PublishedAt:    time.Now(),
		PlatformURL:    fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoResult.ID),
	}, nil
}

func (c *Client) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/channels?part=snippet&mine=true", nil)
	if err != nil {
		return nil, fmt.Errorf("youtube: create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.clientID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("youtube: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("youtube: read body: %w", err)
	}

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				Thumbnails  struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
				} `json:"thumbnails"`
			} `json:"snippet"`
			Statistics struct {
				SubscriberCount int64 `json:"subscriberCount"`
			} `json:"statistics"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("youtube: decode account: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, fmt.Errorf("youtube: no channel found")
	}

	item := result.Items[0]
	return &platform.Account{
		ID:             item.ID,
		PlatformName:   "youtube",
		PlatformUserID: item.ID,
		Name:           item.Snippet.Title,
		Bio:            item.Snippet.Description,
		Avatar:         item.Snippet.Thumbnails.Default.URL,
		FollowerCount:  item.Statistics.SubscriberCount,
	}, nil
}

func (c *Client) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/videos?part=status&id="+platformPostID, nil)
	if err != nil {
		return nil, fmt.Errorf("youtube: create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("youtube: get status: %w", err)
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
		MaxVideoDuration:      43200,
		MaxVideoSize:          256 * 1024 * 1024 * 1024,
		MaxTextLength:         5000,
		MaxTitleLength:        100,
		MaxTagsCount:          15,
		SupportsSchedule:      true,
		SupportsMultiAccount:  true,
		SupportedAspectRatios: []string{"16:9", "4:3"},
	}
}
