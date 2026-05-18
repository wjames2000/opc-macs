// Package platform provides a unified SDK interface for multiple social media platforms.
// It defines the PlatformClient interface that all platform adapters must implement,
// along with shared types, OAuth2.0 flow, token management, and auth adapters.
package platform

import (
	"context"
	"time"
)

// PublishStatus represents the status of a published post.
type PublishStatus string

const (
	PublishStatusPending   PublishStatus = "pending"
	PublishStatusPublished PublishStatus = "published"
	PublishStatusFailed    PublishStatus = "failed"
	PublishStatusPartially PublishStatus = "partially" // some media succeeded, some failed
	PublishStatusScheduled PublishStatus = "scheduled"
)

// PlatformClient defines the interface for all social media platform SDKs.
// Each platform adapter (Douyin, Xiaohongshu, Bilibili, etc.) must implement this interface.
type PlatformClient interface {
	// Name returns the platform name (e.g., "douyin", "xiaohongshu", "youtube").
	Name() string

	// AuthType returns the authentication type used by this platform.
	AuthType() AuthType

	// GetAuthURL generates the OAuth authorization URL.
	// Returns empty string if the platform doesn't support OAuth.
	GetAuthURL(redirectURI, state string) (string, error)

	// ExchangeCode exchanges an OAuth authorization code for an access token.
	ExchangeCode(ctx context.Context, code, redirectURI string) (*AuthToken, error)

	// RefreshToken refreshes an expired access token.
	RefreshToken(ctx context.Context, token *AuthToken) (*AuthToken, error)

	// ValidateToken checks whether the current token is still valid.
	ValidateToken(ctx context.Context, token *AuthToken) (bool, error)

	// Publish publishes content to the platform.
	// Returns a PublishResponse with the platform's post ID and status.
	Publish(ctx context.Context, req *PublishRequest) (*PublishResponse, error)

	// GetAccountInfo retrieves the authenticated account's profile information.
	GetAccountInfo(ctx context.Context) (*Account, error)

	// GetPublishStatus checks the status of a previously published post.
	GetPublishStatus(ctx context.Context, platformPostID string) (*PublishStatusInfo, error)

	// ListAccounts lists accounts managed under this platform client (for multi-account platforms).
	ListAccounts(ctx context.Context) ([]*Account, error)

	// Capabilities returns the publishing capabilities supported by this platform.
	Capabilities() Capabilities
}

// Account represents a social media account bound to the platform.
type Account struct {
	ID             string         `json:"id"`
	PlatformName   string         `json:"platform_name"`
	PlatformUserID string         `json:"platform_user_id"`
	Name           string         `json:"name"`
	Avatar         string         `json:"avatar"`
	Bio            string         `json:"bio"`
	FollowerCount  int64          `json:"follower_count"`
	PlatformExtra  map[string]any `json:"platform_extra,omitempty"`
}

// AuthToken represents an OAuth2.0 or API key authentication token.
type AuthToken struct {
	AccessToken      string         `json:"access_token"`
	RefreshToken     string         `json:"refresh_token,omitempty"`
	ExpiresAt        time.Time      `json:"expires_at"`
	RefreshExpiresAt time.Time      `json:"refresh_expires_at,omitempty"`
	TokenType        string         `json:"token_type"` // Bearer, MAC, etc.
	Scope            string         `json:"scope,omitempty"`
	Raw              map[string]any `json:"raw,omitempty"` // Raw provider-specific fields
}

// IsExpired checks if the token is expired with a 5-minute buffer.
func (t *AuthToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt.Add(-5 * time.Minute))
}

// ContentType represents the type of content being published.
type ContentType string

const (
	ContentTypeVideo   ContentType = "video"
	ContentTypeImage   ContentType = "image"
	ContentTypeText    ContentType = "text"
	ContentTypeArticle ContentType = "article"
	ContentTypeMixed   ContentType = "mixed" // image + text
)

// Content represents the content to be published.
type Content struct {
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Body         string         `json:"body"` // Main text/content
	ContentType  ContentType    `json:"content_type"`
	Tags         []string       `json:"tags,omitempty"`
	MentionUsers []string       `json:"mention_users,omitempty"`
	Location     string         `json:"location,omitempty"`
	Extra        map[string]any `json:"extra,omitempty"` // Platform-specific fields
}

// MediaFile represents a media file to be uploaded.
type MediaFile struct {
	URL       string `json:"url,omitempty"`       // Remote URL to download
	FilePath  string `json:"file_path,omitempty"` // Local file path
	MediaType string `json:"media_type"`          // "video", "image", "audio"
	MimeType  string `json:"mime_type,omitempty"`
	Title     string `json:"title,omitempty"`
	AltText   string `json:"alt_text,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Duration  int    `json:"duration,omitempty"` // Duration in seconds for video/audio
}

// PublishRequest contains all parameters for publishing content to a platform.
type PublishRequest struct {
	AccountID        string         `json:"account_id"`
	AuthToken        *AuthToken     `json:"-"`
	Content          Content        `json:"content"`
	MediaFiles       []MediaFile    `json:"media_files,omitempty"`
	ScheduleAt       *time.Time     `json:"schedule_at,omitempty"`
	PlatformSpecific map[string]any `json:"platform_specific,omitempty"`
}

// PublishResponse contains the result of a publish operation.
type PublishResponse struct {
	PlatformPostID string         `json:"platform_post_id"`
	Status         PublishStatus  `json:"status"`
	PublishedAt    time.Time      `json:"published_at"`
	PostedAt       time.Time      `json:"posted_at,omitempty"`
	PlatformURL    string         `json:"platform_url,omitempty"`
	PostURL        string         `json:"post_url,omitempty"`
	ErrorMessage   string         `json:"error_message,omitempty"`
	Raw            map[string]any `json:"raw,omitempty"`
}

// PublishStatusInfo contains detailed status of a published post.
type PublishStatusInfo struct {
	PlatformPostID string        `json:"platform_post_id"`
	Status         PublishStatus `json:"status"`
	ViewCount      int64         `json:"view_count,omitempty"`
	LikeCount      int64         `json:"like_count,omitempty"`
	CommentCount   int64         `json:"comment_count,omitempty"`
	ShareCount     int64         `json:"share_count,omitempty"`
	ErrorMessage   string        `json:"error_message,omitempty"`
}

// Capabilities describes what a platform can do.
type Capabilities struct {
	SupportedContentTypes []ContentType `json:"supported_content_types"`
	MaxVideoDuration      int           `json:"max_video_duration"`                // seconds, 0 = unlimited
	MaxVideoSize          int64         `json:"max_video_size"`                    // bytes
	MaxImageCount         int           `json:"max_image_count"`                   // 0 = unlimited
	MaxImageSize          int64         `json:"max_image_size"`                    // bytes
	MaxTextLength         int           `json:"max_text_length"`                   // characters
	MaxTitleLength        int           `json:"max_title_length"`                  // characters
	MaxTagsCount          int           `json:"max_tags_count"`                    // 0 = not supported
	SupportsSchedule      bool          `json:"supports_schedule"`                 // scheduled publishing
	SupportsMultiAccount  bool          `json:"supports_multi_account"`            // multi-account management
	SupportedAspectRatios []string      `json:"supported_aspect_ratios,omitempty"` // e.g., "16:9", "9:16", "1:1"
}

// AuthType represents the authentication method used by a platform.
type AuthType string

const (
	AuthTypeOAuth2 AuthType = "oauth2"
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeCookie AuthType = "cookie"
	AuthTypeNone   AuthType = "none"
)
