package linkedin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wjames2000/opc-macs/internal/platform"
)

const (
	baseURL    = "https://api.linkedin.com"
	apiVersion = "/v2"
)

type LinkedinClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func New(clientID, clientSecret string) *LinkedinClient {
	return &LinkedinClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *LinkedinClient) Name() string { return "linkedin" }

func (c *LinkedinClient) AuthType() platform.AuthType { return platform.AuthTypeOAuth2 }

func (c *LinkedinClient) GetAuthURL(redirectURI, state string) (string, error) {
	return fmt.Sprintf(
		"https://www.linkedin.com/oauth/v2/authorization?response_type=code&client_id=%s&redirect_uri=%s&state=%s&scope=w_member_social,openid,profile",
		c.clientID, redirectURI, state), nil
}

func (c *LinkedinClient) ExchangeCode(ctx context.Context, code, redirectURI string) (*platform.AuthToken, error) {
	body := map[string]string{
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  redirectURI,
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequestWithContext(ctx, "POST", "https://www.linkedin.com/oauth/v2/accessToken", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("linkedin: exchange code: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Scope       string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("linkedin: decode token: %w", err)
	}
	if result.AccessToken == "" {
		return nil, fmt.Errorf("linkedin: empty access_token in response")
	}

	return &platform.AuthToken{
		AccessToken: result.AccessToken,
		ExpiresAt:   time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
		Scope:       result.Scope,
	}, nil
}

func (c *LinkedinClient) RefreshToken(ctx context.Context, token *platform.AuthToken) (*platform.AuthToken, error) {
	return nil, fmt.Errorf("linkedin: refresh token not supported via API, user must re-authenticate")
}

func (c *LinkedinClient) ValidateToken(ctx context.Context, token *platform.AuthToken) (bool, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/me", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200, nil
}

func (c *LinkedinClient) Publish(ctx context.Context, req *platform.PublishRequest) (*platform.PublishResponse, error) {
	if req.AuthToken == nil || req.AuthToken.AccessToken == "" {
		return nil, fmt.Errorf("linkedin: no access token provided")
	}

	author, err := c.resolveAuthor(ctx, req.AuthToken.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("linkedin: resolve author: %w", err)
	}

	postBody := map[string]interface{}{
		"author":         "urn:li:person:" + author,
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]interface{}{
				"shareCommentary": map[string]string{
					"text": req.Content.Body,
				},
				"shareMediaCategory": "NONE",
			},
		},
		"visibility": map[string]string{
			"com.linkedin.ugc.MemberNetworkVisibility": "PUBLIC",
		},
	}

	if len(req.MediaFiles) > 0 {
		url, _, _ := strings.Cut(req.MediaFiles[0].URL, "?")
		if url == "" {
			url = req.MediaFiles[0].FilePath
		}
		mediaCategory := "IMAGE"
		if req.MediaFiles[0].MediaType == "video" {
			mediaCategory = "VIDEO"
		}
		postBody["specificContent"].(map[string]interface{})["com.linkedin.ugc.ShareContent"].(map[string]interface{})["shareMediaCategory"] = mediaCategory
		postBody["specificContent"].(map[string]interface{})["com.linkedin.ugc.ShareContent"].(map[string]interface{})["media"] = []map[string]interface{}{
			{"status": "READY", "media": url},
		}
	}

	data, _ := json.Marshal(postBody)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", baseURL+apiVersion+"/ugcPosts", bytes.NewReader(data))
	httpReq.Header.Set("Authorization", "Bearer "+req.AuthToken.AccessToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Restli-Protocol-Version", "2.0.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("linkedin: publish: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("linkedin: publish failed (%d): %s", resp.StatusCode, string(respBody))
	}

	location := resp.Header.Get("X-RestLi-Id")
	return &platform.PublishResponse{
		PlatformPostID: location,
		Status:         platform.PublishStatusPublished,
	}, nil
}

func (c *LinkedinClient) GetPublishStatus(ctx context.Context, platformPostID string) (*platform.PublishStatusInfo, error) {
	return &platform.PublishStatusInfo{
		Status:         platform.PublishStatusPublished,
		PlatformPostID: platformPostID,
	}, nil
}

func (c *LinkedinClient) ListAccounts(ctx context.Context) ([]*platform.Account, error) {
	acct, err := c.GetAccountInfo(ctx)
	if err != nil {
		return nil, err
	}
	return []*platform.Account{acct}, nil
}

func (c *LinkedinClient) Capabilities() platform.Capabilities {
	return platform.Capabilities{
		SupportedContentTypes: []platform.ContentType{platform.ContentTypeText, platform.ContentTypeImage, platform.ContentTypeArticle},
		MaxTextLength:         3000,
		MaxImageCount:         9,
		MaxTitleLength:        200,
		MaxTagsCount:          0,
		SupportsSchedule:      false,
		SupportsMultiAccount:  false,
	}
}

func (c *LinkedinClient) resolveAuthor(ctx context.Context, accessToken string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var me struct {
		Sub string `json:"sub"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
		return "", err
	}
	return me.Sub, nil
}

func (c *LinkedinClient) GetAccountInfo(ctx context.Context) (*platform.Account, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", baseURL+apiVersion+"/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("linkedin: create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.clientID)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("linkedin: get account: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("linkedin: read body: %w", err)
	}

	var result struct {
		Sub     string `json:"sub"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
		Locale  string `json:"locale"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("linkedin: decode account: %w", err)
	}

	return &platform.Account{
		ID:             result.Sub,
		PlatformName:   "linkedin",
		PlatformUserID: result.Sub,
		Name:           result.Name,
		Avatar:         result.Picture,
		Bio:            "",
	}, nil
}
