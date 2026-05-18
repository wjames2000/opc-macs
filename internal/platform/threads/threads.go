package threads

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
	baseURL = "https://graph.threads.net"
	version = "/v1.0"
)

type ThreadsClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func New(clientID, clientSecret string) *ThreadsClient {
	return &ThreadsClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *ThreadsClient) Name() string { return "threads" }

func (c *ThreadsClient) AuthType() platform.AuthType { return platform.AuthTypeOAuth2 }

func (c *ThreadsClient) GetAuthURL(redirectURI, state string) (string, error) {
	return fmt.Sprintf(
		"https://www.threads.net/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s&scope=threads_basic,threads_content_publish&response_type=code",
		c.clientID, redirectURI, state), nil
}

func (c *ThreadsClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*platform.AuthToken, error) {
	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	form.Set("grant_type", "authorization_code")

	req, _ := http.NewRequestWithContext(ctx, "POST", baseURL+version+"/oauth/access_token", bytes.NewReader([]byte(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("threads: exchange code: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int64  `json:"expires_in"`
		UserID      string `json:"user_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("threads: decode token: %w", err)
	}
	if result.AccessToken == "" {
		return nil, fmt.Errorf("threads: empty access_token")
	}

	return &platform.AuthToken{
		AccessToken: result.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
		Raw:         map[string]any{"user_id": result.UserID},
	}, nil
}

func (c *ThreadsClient) RefreshToken(ctx context.Context, token *platform.AuthToken) (*platform.AuthToken, error) {
	form := url.Values{}
	form.Set("grant_type", "th_refresh_token")
	form.Set("access_token", token.AccessToken)

	req, _ := http.NewRequestWithContext(ctx, "POST", baseURL+version+"/oauth/refresh_access_token", bytes.NewReader([]byte(form.Encode())))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("threads: refresh token: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("threads: decode refresh: %w", err)
	}

	return &platform.AuthToken{
		AccessToken: result.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
	}, nil
}

func (c *ThreadsClient) ValidateToken(ctx context.Context, token *platform.AuthToken) (bool, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", baseURL+version+"/me?fields=id&access_token="+token.AccessToken, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200, nil
}

func (c *ThreadsClient) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("threads: no access token")
	}

	userID, _ := req.AuthToken.Raw["user_id"].(string)
	if userID == "" {
		return nil, fmt.Errorf("threads: user_id not found in token")
	}
	createURL := fmt.Sprintf("%s%s/%s/threads?access_token=%s",
		baseURL, version, userID, req.AuthToken.AccessToken)

	createBody := map[string]string{
		"text":       req.Content.Body,
		"media_type": "TEXT",
	}
	if len(req.MediaFiles) > 0 {
		createBody["media_type"] = "IMAGE"
		createBody["image_url"] = req.MediaFiles[0].URL
	}
	data, _ := json.Marshal(createBody)

	createResp, err := c.httpClient.Post(createURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("threads: create: %w", err)
	}
	defer createResp.Body.Close()

	var creation struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&creation); err != nil {
		return nil, fmt.Errorf("threads: decode creation: %w", err)
	}
	if creation.ID == "" {
		return nil, fmt.Errorf("threads: empty media container id")
	}

	publishURL := fmt.Sprintf("%s%s/%s/threads_publish?access_token=%s&creation_id=%s",
		baseURL, version, userID, req.AuthToken.AccessToken, creation.ID)

	pubResp, err := c.httpClient.Post(publishURL, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("threads: publish: %w", err)
	}
	defer pubResp.Body.Close()

	if pubResp.StatusCode != 200 {
		body, _ := io.ReadAll(pubResp.Body)
		return nil, fmt.Errorf("threads: publish failed (%d): %s", pubResp.StatusCode, string(body))
	}

	var publishResult struct {
		ID string `json:"id"`
	}
	json.NewDecoder(pubResp.Body).Decode(&publishResult)

	return &platform.PublishResponse{
		PlatformPostID: publishResult.ID,
		Status:         platform.PublishStatusPublished,
	}, nil
}

func (c *ThreadsClient) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	getURL := fmt.Sprintf("%s%s/me?fields=id,name,username,threads_profile_picture_url&access_token=%s",
		baseURL, version, c.clientID)

	resp, err := c.httpClient.Get(getURL)
	if err != nil {
		return nil, fmt.Errorf("threads: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("threads: read body: %w", err)
	}

	var result struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
		Picture  string `json:"threads_profile_picture_url"`
		Error    *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("threads: decode account: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("threads: get account failed: %s", result.Error.Message)
	}

	return &platform.Account{
		ID:             result.ID,
		PlatformName:   "threads",
		PlatformUserID: result.Username,
		Name:           result.Name,
		Avatar:         result.Picture,
	}, nil
}

func (c *ThreadsClient) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	return &platform.PublishStatusInfo{
		Status:         platform.PublishStatusPublished,
		PlatformPostID: platformPostID,
	}, nil
}

func (c *ThreadsClient) ListAccounts(ctx context.Context) ([]*platform.Account, error) {
	acct, err := c.GetAccountInfo(ctx)
	if err != nil {
		return nil, err
	}
	return []*platform.Account{acct}, nil
}

func (c *ThreadsClient) Capabilities() platform.Capabilities {
	return platform.Capabilities{
		SupportedContentTypes: []platform.ContentType{platform.ContentTypeText, platform.ContentTypeImage},
		MaxTextLength:         500,
		MaxImageCount:         1,
		MaxTagsCount:          0,
		SupportsSchedule:      false,
		SupportsMultiAccount:  false,
	}
}
