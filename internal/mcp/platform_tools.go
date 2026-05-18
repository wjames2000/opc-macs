package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/wjames2000/opc-macs/internal/platform"
)

// PlatformToolSet registers platform-related tools with the MCP server.
type PlatformToolSet struct {
	clientProvider func(ctx context.Context, platformName string) (platform.PlatformClient, error)
	tokenProvider  func(ctx context.Context, accountID string) (*platform.AuthToken, error)
}

// NewPlatformToolSet creates a new PlatformToolSet with the given provider functions.
func NewPlatformToolSet(
	clientProvider func(ctx context.Context, platformName string) (platform.PlatformClient, error),
	tokenProvider func(ctx context.Context, accountID string) (*platform.AuthToken, error),
) *PlatformToolSet {
	return &PlatformToolSet{
		clientProvider: clientProvider,
		tokenProvider:  tokenProvider,
	}
}

// RegisterAll registers all platform tools with the MCP server.
func (pts *PlatformToolSet) RegisterAll(s *Server) {
	s.RegisterTool(
		"bind_platform_account",
		"Bind a social media platform account via OAuth2. Returns the authorization URL that the user must visit to complete the binding.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"platform": map[string]any{
					"type":        "string",
					"description": "Platform name: douyin, xiaohongshu, bilibili, kuaishou, wechat, youtube, tiktok, facebook, instagram, twitter, pinterest, linkedin, threads",
				},
				"redirect_uri": map[string]any{
					"type":        "string",
					"description": "OAuth redirect URI",
				},
			},
			"required": []string{"platform", "redirect_uri"},
		},
		pts.handleBindAccount,
	)

	s.RegisterTool(
		"list_platform_accounts",
		"List all bound social media platform accounts with their status.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"platform": map[string]any{
					"type":        "string",
					"description": "Optional platform filter (e.g., 'douyin'). Empty returns all.",
				},
			},
		},
		pts.handleListAccounts,
	)

	s.RegisterTool(
		"get_publish_capabilities",
		"Get the publishing capabilities of a specific platform (supported content types, limits, etc.).",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"platform": map[string]any{
					"type":        "string",
					"description": "Platform name to check capabilities for",
				},
			},
			"required": []string{"platform"},
		},
		pts.handleGetCapabilities,
	)

	s.RegisterTool(
		"publish_content",
		"Publish content to a bound platform account. Supports video, image, and text content types. Content is auto-adapted to platform-specific requirements.",
		map[string]any{
			"type": "object",
			"properties": map[string]any{
				"account_id": map[string]any{
					"type":        "string",
					"description": "The bound account ID to publish to",
				},
				"platform": map[string]any{
					"type":        "string",
					"description": "Platform name",
				},
				"content_type": map[string]any{
					"type":        "string",
					"description": "Content type: video, image, text, mixed",
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Content title (auto-truncated to platform limit)",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "Main text content (auto-truncated to platform limit)",
				},
				"media_urls": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "URLs of media files to include",
				},
				"tags": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Hashtags (auto-converted to platform format)",
				},
			},
			"required": []string{"account_id", "platform", "content_type"},
		},
		pts.handlePublish,
	)
}

func (pts *PlatformToolSet) handleBindAccount(ctx context.Context, params json.RawMessage) (any, error) {
	var req struct {
		Platform    string `json:"platform"`
		RedirectURI string `json:"redirect_uri"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	client, err := pts.clientProvider(ctx, req.Platform)
	if err != nil {
		return nil, fmt.Errorf("get platform client: %w", err)
	}

	authURL, err := client.GetAuthURL(req.RedirectURI, "")
	if err != nil {
		return nil, fmt.Errorf("get auth URL: %w", err)
	}

	return map[string]any{
		"platform": req.Platform,
		"auth_url": authURL,
		"instructions": fmt.Sprintf(
			"Visit the following URL to authorize OPC-Agent to access your %s account:\n%s\n\nAfter authorization, you will be redirected to a callback URL with an authorization code. Use the 'complete_bind_platform_account' tool to complete the binding.",
			req.Platform, authURL,
		),
	}, nil
}

func (pts *PlatformToolSet) handleListAccounts(ctx context.Context, params json.RawMessage) (any, error) {
	type listReq struct {
		Platform string `json:"platform"`
	}
	var req listReq
	json.Unmarshal(params, &req)

	// In a real implementation, this would query the TokenStore
	// For now return the tool interface description
	return map[string]any{
		"message": "List accounts - requires TokenStore implementation. See WBS 1.4.",
		"platforms": []string{
			"douyin", "xiaohongshu", "bilibili", "kuaishou", "wechat",
			"youtube", "tiktok", "facebook", "instagram", "twitter",
			"pinterest", "linkedin", "threads",
		},
	}, nil
}

func (pts *PlatformToolSet) handleGetCapabilities(ctx context.Context, params json.RawMessage) (any, error) {
	var req struct {
		Platform string `json:"platform"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	client, err := pts.clientProvider(ctx, req.Platform)
	if err != nil {
		return nil, fmt.Errorf("get platform client: %w", err)
	}

	return client.Capabilities(), nil
}

func (pts *PlatformToolSet) handlePublish(ctx context.Context, params json.RawMessage) (any, error) {
	var req struct {
		AccountID   string   `json:"account_id"`
		Platform    string   `json:"platform"`
		ContentType string   `json:"content_type"`
		Title       string   `json:"title"`
		Body        string   `json:"body"`
		MediaURLs   []string `json:"media_urls"`
		Tags        []string `json:"tags"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, fmt.Errorf("invalid params: %w", err)
	}

	client, err := pts.clientProvider(ctx, req.Platform)
	if err != nil {
		return nil, fmt.Errorf("get platform client: %w", err)
	}

	pubReq := &platform.PublishRequest{
		AccountID: req.AccountID,
		Content: platform.Content{
			Title:       req.Title,
			Body:        req.Body,
			ContentType: platform.ContentType(req.ContentType),
			Tags:        req.Tags,
		},
	}

	for _, url := range req.MediaURLs {
		pubReq.MediaFiles = append(pubReq.MediaFiles, platform.MediaFile{
			URL:       url,
			MediaType: req.ContentType,
		})
	}

	resp, err := client.Publish(ctx, pubReq)
	if err != nil {
		return nil, fmt.Errorf("publish failed: %w", err)
	}

	log.Printf("[MCP] Published content to %s: post_id=%s, status=%s",
		req.Platform, resp.PlatformPostID, resp.Status)

	return resp, nil
}
