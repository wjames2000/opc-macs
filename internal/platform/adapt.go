package platform

import (
	"fmt"
	"strings"
)

// AspectAdapter handles video/image aspect ratio conversion suggestions.
// When content needs to be published across platforms with different ratio requirements,
// this adapter provides the optimal crop/format suggestions.
type AspectAdapter struct {
	ratioConfig map[string]map[ContentType]string
}

func NewAspectAdapter() *AspectAdapter {
	return &AspectAdapter{
		ratioConfig: map[string]map[ContentType]string{
			"douyin":      {ContentTypeVideo: "9:16", ContentTypeImage: "9:16"},
			"xiaohongshu": {ContentTypeVideo: "3:4", ContentTypeImage: "3:4"},
			"bilibili":    {ContentTypeVideo: "16:9", ContentTypeImage: "16:9"},
			"kuaishou":    {ContentTypeVideo: "9:16", ContentTypeImage: "9:16"},
			"youtube":     {ContentTypeVideo: "16:9", ContentTypeImage: "16:9"},
			"tiktok":      {ContentTypeVideo: "9:16", ContentTypeImage: "1:1"},
			"instagram":   {ContentTypeVideo: "9:16", ContentTypeImage: "1:1"},
			"facebook":    {ContentTypeVideo: "16:9", ContentTypeImage: "1.91:1"},
			"twitter":     {ContentTypeVideo: "16:9", ContentTypeImage: "16:9"},
			"linkedin":    {ContentTypeImage: "1.91:1", ContentTypeVideo: "16:9"},
			"pinterest":   {ContentTypeImage: "2:3", ContentTypeVideo: "2:3"},
		},
	}
}

func (aa *AspectAdapter) Suggest(platformName string, contentType ContentType) string {
	if ratios, ok := aa.ratioConfig[platformName]; ok {
		if ratio, ok := ratios[contentType]; ok {
			return ratio
		}
	}
	return "16:9"
}

func (aa *AspectAdapter) SuggestAll(platformName string) []string {
	ratios, ok := aa.ratioConfig[platformName]
	if !ok {
		return []string{"16:9"}
	}
	seen := make(map[string]bool)
	var result []string
	for _, ratio := range ratios {
		if !seen[ratio] {
			seen[ratio] = true
			result = append(result, ratio)
		}
	}
	return result
}

// LengthAdapter handles text length adaptation across platforms.
// It ensures content fits within each platform's character limits.
type LengthAdapter struct {
	limits map[string]TextLimit
}

type TextLimit struct {
	Title int
	Body  int
	Bio   int
}

func NewLengthAdapter() *LengthAdapter {
	return &LengthAdapter{
		limits: map[string]TextLimit{
			"douyin":      {Title: 30, Body: 1000},
			"xiaohongshu": {Title: 20, Body: 2000},
			"bilibili":    {Title: 50, Body: 3000},
			"kuaishou":    {Title: 25, Body: 500},
			"wechat":      {Title: 64, Body: 50000},
			"youtube":     {Title: 100, Body: 5000},
			"tiktok":      {Title: 50, Body: 4000},
			"facebook":    {Title: 80, Body: 63206},
			"instagram":   {Title: 0, Body: 2200},
			"twitter":     {Title: 0, Body: 280},
			"linkedin":    {Title: 70, Body: 3000},
			"pinterest":   {Title: 100, Body: 500},
		},
	}
}

func (la *LengthAdapter) AdaptTitle(platformName, title string) string {
	limit, ok := la.limits[platformName]
	if !ok || limit.Title == 0 {
		return title
	}
	runes := []rune(title)
	if len(runes) > limit.Title {
		return string(runes[:limit.Title])
	}
	return title
}

func (la *LengthAdapter) AdaptBody(platformName, body string) string {
	limit, ok := la.limits[platformName]
	if !ok || limit.Body == 0 {
		return body
	}
	runes := []rune(body)
	if len(runes) > limit.Body {
		return string(runes[:limit.Body])
	}
	return body
}

func (la *LengthAdapter) GetLimits(platformName string) (TextLimit, error) {
	limit, ok := la.limits[platformName]
	if !ok {
		return TextLimit{}, fmt.Errorf("no limits defined for platform: %s", platformName)
	}
	return limit, nil
}

// TagAdapter handles tag format conversion across platforms.
type TagAdapter struct {
	rules map[string]TagRule
}

type TagRule struct {
	Prefix        string
	Suffix        string
	MaxTags       int
	MaxTagLen     int
	AllowSpaces   bool
	CaseSensitive bool
}

func NewTagAdapter() *TagAdapter {
	return &TagAdapter{
		rules: map[string]TagRule{
			"douyin":      {Prefix: "#", MaxTags: 10, MaxTagLen: 20, AllowSpaces: false, CaseSensitive: false},
			"xiaohongshu": {Prefix: "#", MaxTags: 10, MaxTagLen: 20, AllowSpaces: false, CaseSensitive: false},
			"bilibili":    {Prefix: "#", MaxTags: 6, MaxTagLen: 30, AllowSpaces: false, CaseSensitive: false},
			"kuaishou":    {Prefix: "#", MaxTags: 8, MaxTagLen: 15, AllowSpaces: false, CaseSensitive: false},
			"youtube":     {Prefix: "#", MaxTags: 15, MaxTagLen: 30, AllowSpaces: true, CaseSensitive: true},
			"tiktok":      {Prefix: "#", MaxTags: 10, MaxTagLen: 30, AllowSpaces: false, CaseSensitive: true},
			"facebook":    {Prefix: "#", MaxTags: 5, MaxTagLen: 30, AllowSpaces: true, CaseSensitive: true},
			"instagram":   {Prefix: "#", MaxTags: 30, MaxTagLen: 30, AllowSpaces: false, CaseSensitive: true},
			"twitter":     {Prefix: "#", MaxTags: 0, MaxTagLen: 30, AllowSpaces: false, CaseSensitive: true},
			"linkedin":    {Prefix: "#", MaxTags: 5, MaxTagLen: 30, AllowSpaces: true, CaseSensitive: true},
			"pinterest":   {Prefix: "", MaxTags: 20, MaxTagLen: 30, AllowSpaces: true, CaseSensitive: true},
		},
	}
}

func (ta *TagAdapter) Convert(platformName string, tags []string) []string {
	rule, ok := ta.rules[platformName]
	if !ok {
		return tags
	}
	var result []string
	for _, tag := range tags {
		if rule.MaxTags > 0 && len(result) >= rule.MaxTags {
			break
		}
		cleanTag := strings.TrimSpace(tag)
		cleanTag = strings.TrimLeft(cleanTag, "#@")
		if !rule.AllowSpaces {
			cleanTag = strings.ReplaceAll(cleanTag, " ", "_")
		}
		if rule.MaxTagLen > 0 && len([]rune(cleanTag)) > rule.MaxTagLen {
			cleanTag = string([]rune(cleanTag)[:rule.MaxTagLen])
		}
		if rule.Prefix != "" {
			cleanTag = rule.Prefix + cleanTag
		}
		if rule.Suffix != "" {
			cleanTag = cleanTag + rule.Suffix
		}
		result = append(result, cleanTag)
	}
	return result
}
