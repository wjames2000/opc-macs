package publish

import (
	"strings"

	"github.com/wjames2000/opc-macs/internal/platform"
)

// ContentAdapter handles cross-platform content adaptation.
// It automatically adjusts content (aspect ratio, text length, tag format)
// to meet each platform's specific requirements.
type ContentAdapter struct {
	adapters map[string]PlatformAdapter
}

// PlatformAdapter defines per-platform content transformation rules.
type PlatformAdapter struct {
	MaxTextLength  int
	MaxTitleLength int
	MaxTags        int
	TagPrefix      string // "#" for most, "@" for some
	TagSuffix      string
	Separator      string
}

// NewContentAdapter creates a ContentAdapter with default platform rules.
func NewContentAdapter() *ContentAdapter {
	return &ContentAdapter{
		adapters: map[string]PlatformAdapter{
			"douyin":      {MaxTextLength: 1000, MaxTitleLength: 30, MaxTags: 10, TagPrefix: "#"},
			"xiaohongshu": {MaxTextLength: 2000, MaxTitleLength: 20, MaxTags: 10, TagPrefix: "#"},
			"bilibili":    {MaxTextLength: 3000, MaxTitleLength: 50, MaxTags: 6, TagPrefix: "#"},
			"kuaishou":    {MaxTextLength: 500, MaxTitleLength: 25, MaxTags: 8, TagPrefix: "#"},
			"wechat":      {MaxTextLength: 50000, MaxTitleLength: 64, MaxTags: 0, TagPrefix: "#"},
			"youtube":     {MaxTextLength: 5000, MaxTitleLength: 100, MaxTags: 15, TagPrefix: "#"},
			"tiktok":      {MaxTextLength: 4000, MaxTitleLength: 50, MaxTags: 10, TagPrefix: "#"},
			"facebook":    {MaxTextLength: 63206, MaxTitleLength: 80, MaxTags: 5, TagPrefix: "#"},
			"instagram":   {MaxTextLength: 2200, MaxTitleLength: 0, MaxTags: 30, TagPrefix: "#"},
			"twitter":     {MaxTextLength: 280, MaxTitleLength: 0, MaxTags: 0, TagPrefix: "#"},
			"linkedin":    {MaxTextLength: 3000, MaxTitleLength: 70, MaxTags: 5, TagPrefix: "#"},
			"pinterest":   {MaxTextLength: 500, MaxTitleLength: 100, MaxTags: 20, TagPrefix: ""},
		},
	}
}

// RegisterAdapter registers or overrides a platform's adapter rules.
func (ca *ContentAdapter) RegisterAdapter(platform string, adapter PlatformAdapter) {
	ca.adapters[platform] = adapter
}

// AdaptTitle truncates the title to the platform's maximum length.
func (ca *ContentAdapter) AdaptTitle(platformName, title string) string {
	adapter, ok := ca.adapters[platformName]
	if !ok || adapter.MaxTitleLength == 0 {
		return title
	}
	runes := []rune(title)
	if len(runes) > adapter.MaxTitleLength {
		return string(runes[:adapter.MaxTitleLength])
	}
	return title
}

// AdaptBody truncates the body text to the platform's maximum length.
func (ca *ContentAdapter) AdaptBody(platformName, body string) string {
	adapter, ok := ca.adapters[platformName]
	if !ok || adapter.MaxTextLength == 0 {
		return body
	}
	runes := []rune(body)
	if len(runes) > adapter.MaxTextLength {
		return string(runes[:adapter.MaxTextLength])
	}
	return body
}

// AdaptTags converts tags to the platform's format (prefix, suffix, count limit).
func (ca *ContentAdapter) AdaptTags(platformName string, tags []string) []string {
	adapter, ok := ca.adapters[platformName]
	if !ok {
		return tags
	}

	var result []string
	for _, tag := range tags {
		if adapter.MaxTags > 0 && len(result) >= adapter.MaxTags {
			break
		}
		cleanTag := strings.TrimSpace(tag)
		cleanTag = strings.TrimLeft(cleanTag, "#@")
		if adapter.TagPrefix != "" {
			cleanTag = adapter.TagPrefix + cleanTag
		}
		if adapter.TagSuffix != "" {
			cleanTag = cleanTag + adapter.TagSuffix
		}
		result = append(result, cleanTag)
	}
	return result
}

// AdaptContent applies all content adaptations for the given platform.
func (ca *ContentAdapter) AdaptContent(platformName string, content *platform.Content) *platform.Content {
	adapted := *content
	adapted.Title = ca.AdaptTitle(platformName, content.Title)
	adapted.Body = ca.AdaptBody(platformName, content.Body)
	adapted.Tags = ca.AdaptTags(platformName, content.Tags)
	return &adapted
}

// SuggestAspectRatio returns the recommended aspect ratio for a platform.
func (ca *ContentAdapter) SuggestAspectRatio(platformName string, contentType platform.ContentType) string {
	suggestions := map[string]map[platform.ContentType]string{
		"douyin":      {platform.ContentTypeVideo: "9:16", platform.ContentTypeImage: "1:1"},
		"xiaohongshu": {platform.ContentTypeVideo: "3:4", platform.ContentTypeImage: "3:4"},
		"bilibili":    {platform.ContentTypeVideo: "16:9"},
		"kuaishou":    {platform.ContentTypeVideo: "9:16"},
		"youtube":     {platform.ContentTypeVideo: "16:9"},
		"tiktok":      {platform.ContentTypeVideo: "9:16"},
		"instagram":   {platform.ContentTypeImage: "1:1", platform.ContentTypeVideo: "9:16"},
		"facebook":    {platform.ContentTypeVideo: "16:9", platform.ContentTypeImage: "1.91:1"},
		"twitter":     {platform.ContentTypeImage: "16:9"},
		"linkedin":    {platform.ContentTypeImage: "1.91:1"},
		"pinterest":   {platform.ContentTypeImage: "2:3"},
	}

	if ratios, ok := suggestions[platformName]; ok {
		if ratio, ok := ratios[contentType]; ok {
			return ratio
		}
	}
	return "16:9"
}
