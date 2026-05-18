package douyin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/wjames2000/opc-macs/internal/platform"
)

const (
	baseURL    = "https://open.douyin.com"
	apiVersion = "/api/douyin/v1"
)

type DouyinClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func New(clientID, clientSecret string) *DouyinClient {
	return &DouyinClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *DouyinClient) Name() string { return "douyin" }

func (c *DouyinClient) AuthType() platform.AuthType { return platform.AuthTypeOAuth2 }

func (c *DouyinClient) GetAuthURL(redirectURI, state string) (string, error) {
	return fmt.Sprintf("%s/oauth/authorize?client_key=%s&redirect_uri=%s&state=%s&response_type=code",
		baseURL, c.clientID, redirectURI, state), nil
}

func (c *DouyinClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*platform.AuthToken, error) {
	body := map[string]string{
		"client_key":    c.clientID,
		"client_secret": c.clientSecret,
		"code":          code,
		"grant_type":    "authorization_code",
		"redirect_uri":  redirectURI,
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", baseURL+"/oauth/access_token", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("douyin: exchange code: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			AccessToken      string `json:"access_token"`
			ExpiresIn        int64  `json:"expires_in"`
			RefreshToken     string `json:"refresh_token"`
			RefreshExpiresIn int64  `json:"refresh_expires_in"`
			OpenID           string `json:"open_id"`
			UnionID          string `json:"union_id"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("douyin: decode token response: %w", err)
	}
	if result.Data.AccessToken == "" {
		return nil, fmt.Errorf("douyin: token exchange failed: %s", result.Message)
	}

	now := time.Now()
	return &platform.AuthToken{
		AccessToken:      result.Data.AccessToken,
		TokenType:        "Bearer",
		ExpiresAt:        now.Add(time.Duration(result.Data.ExpiresIn) * time.Second),
		RefreshToken:     result.Data.RefreshToken,
		RefreshExpiresAt: now.Add(time.Duration(result.Data.RefreshExpiresIn) * time.Second),
		Scope:            "video.publish,video.upload,user.info",
	}, nil
}

func (c *DouyinClient) RefreshToken(ctx context.Context, token *platform.AuthToken) (*platform.AuthToken, error) {
	body := map[string]string{
		"client_key":    c.clientID,
		"client_secret": c.clientSecret,
		"grant_type":    "refresh_token",
		"refresh_token": token.RefreshToken,
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", baseURL+"/oauth/refresh_token", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("douyin: refresh token: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			AccessToken      string `json:"access_token"`
			ExpiresIn        int64  `json:"expires_in"`
			RefreshToken     string `json:"refresh_token"`
			RefreshExpiresIn int64  `json:"refresh_expires_in"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("douyin: decode refresh: %w", err)
	}

	now := time.Now()
	return &platform.AuthToken{
		AccessToken:      result.Data.AccessToken,
		TokenType:        "Bearer",
		ExpiresAt:        now.Add(time.Duration(result.Data.ExpiresIn) * time.Second),
		RefreshToken:     result.Data.RefreshToken,
		RefreshExpiresAt: now.Add(time.Duration(result.Data.RefreshExpiresIn) * time.Second),
	}, nil
}

func (c *DouyinClient) ValidateToken(ctx context.Context, token *platform.AuthToken) (bool, error) {
	if token == nil || token.AccessToken == "" {
		return false, nil
	}
	if time.Now().After(token.ExpiresAt) {
		return false, nil
	}
	return true, nil
}

func (c *DouyinClient) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil {
		return nil, fmt.Errorf("douyin: auth token required")
	}
	if len(req.MediaFiles) == 0 {
		return nil, fmt.Errorf("douyin: at least one media file required")
	}

	mediaID, err := c.uploadVideo(ctx, req.AuthToken.AccessToken, &req.MediaFiles[0])
	if err != nil {
		return nil, fmt.Errorf("douyin: upload video: %w", err)
	}

	type createReq struct {
		MediaID    string `json:"media_id"`
		Title      string `json:"text"`
		Privacy    int    `json:"privacy"`
		ScheduleTS int64  `json:"schedule_time,omitempty"`
	}
	cr := createReq{
		MediaID: mediaID,
		Title:   req.Content.Title,
		Privacy: 0,
	}
	if !req.ScheduleAt.IsZero() {
		cr.ScheduleTS = req.ScheduleAt.Unix()
	}

	data, _ := json.Marshal(cr)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/video/create/", bytes.NewReader(data))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("access-token", req.AuthToken.AccessToken)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("douyin: create video: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			ItemID string `json:"item_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("douyin: decode create response: %w", err)
	}

	return &platform.PublishResponse{
		PlatformPostID: result.Data.ItemID,
		Status:         platform.PublishStatusPublished,
		PostedAt:       time.Now(),
		PlatformURL:    fmt.Sprintf("https://www.douyin.com/video/%s", result.Data.ItemID),
	}, nil
}

func (c *DouyinClient) uploadVideo(ctx context.Context, accessToken string, media *platform.MediaFile) (string, error) {
	body := map[string]string{"open_id": accessToken}
	data, _ := json.Marshal(body)

	initReq, _ := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/video/upload/init/", bytes.NewReader(data))
	initReq.Header.Set("Content-Type", "application/json")
	initReq.Header.Set("access-token", accessToken)

	resp, err := c.httpClient.Do(initReq)
	if err != nil {
		return "", fmt.Errorf("douyin: init upload: %w", err)
	}
	defer resp.Body.Close()

	var initResult struct {
		Data struct {
			UploadID string `json:"upload_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&initResult); err != nil {
		return "", fmt.Errorf("douyin: decode init: %w", err)
	}

	var fileData []byte
	if media.FilePath != "" {
		fileData, _ = os.ReadFile(media.FilePath)
	}
	uploadReq, _ := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/video/upload/part/", bytes.NewReader(fileData))
	uploadReq.Header.Set("Content-Type", "application/octet-stream")
	uploadReq.Header.Set("access-token", accessToken)
	uploadReq.Header.Set("upload-id", initResult.Data.UploadID)

	uploadResp, err := c.httpClient.Do(uploadReq)
	if err != nil {
		return "", fmt.Errorf("douyin: upload part: %w", err)
	}
	defer uploadResp.Body.Close()
	io.Copy(io.Discard, uploadResp.Body)

	return initResult.Data.UploadID, nil
}

func (c *DouyinClient) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/oauth/userinfo/", nil)
	if err != nil {
		return nil, fmt.Errorf("douyin: create request: %w", err)
	}
	req.Header.Set("access-token", c.clientID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("douyin: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("douyin: read body: %w", err)
	}

	var result struct {
		Data struct {
			OpenID    string `json:"open_id"`
			UnionID   string `json:"union_id"`
			Nickname  string `json:"nickname"`
			AvatarURL string `json:"avatar"`
			Followers int64  `json:"followers"`
		} `json:"data"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("douyin: decode account: %w", err)
	}
	if result.Data.OpenID == "" {
		return nil, fmt.Errorf("douyin: get account failed: %s", result.Message)
	}

	return &platform.Account{
		ID:             result.Data.OpenID,
		PlatformName:   "douyin",
		PlatformUserID: result.Data.OpenID,
		Name:           result.Data.Nickname,
		Avatar:         result.Data.AvatarURL,
		FollowerCount:  result.Data.Followers,
	}, nil
}

func (c *DouyinClient) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/video/status/?item_id="+platformPostID, nil)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("douyin: get publish status: %w", err)
	}
	defer resp.Body.Close()

	return &platform.PublishStatusInfo{
		Status:         platform.PublishStatusPublished,
		PlatformPostID: platformPostID,
	}, nil
}

func (c *DouyinClient) ListAccounts(ctx context.Context) ([]*platform.Account, error) {
	acct, err := c.GetAccountInfo(ctx)
	if err != nil {
		return nil, err
	}
	return []*platform.Account{acct}, nil
}

func (c *DouyinClient) Capabilities() platform.Capabilities {
	return platform.Capabilities{
		SupportedContentTypes: []platform.ContentType{platform.ContentTypeVideo},
		MaxVideoDuration:      900,
		MaxVideoSize:          4 * 1024 * 1024 * 1024,
		MaxImageCount:         0,
		MaxTextLength:         1000,
		MaxTitleLength:        30,
		MaxTagsCount:          10,
		SupportsSchedule:      true,
		SupportsMultiAccount:  false,
		SupportedAspectRatios: []string{"9:16"},
	}
}
