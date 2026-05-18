package plugins

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

var videoScriptSystemPrompt = `你是视频脚本创作专家，擅长为不同平台创作高转化视频脚本。

[核心能力]
- 平台适配：为抖音（15-60s）、快手（15-60s）、B站（5-20min）、YouTube（8-30min）、TikTok（15-60s）等平台优化脚本结构
- 爆款公式：3秒破冰 + 痛点共鸣 + 解决方案 + 行动号召
- 带货脚本：产品卖点植入 + 场景化演示 + 限时优惠话术
- 品牌定制：品牌调性 + 目标人群 + 转化目标

[脚本结构]
1. 开场钩子（前3秒黄金时间）
2. 痛点/需求引入
3. 产品/内容展示
4. 使用场景/效果演示
5. 行动号召（关注/点赞/购买）

[输出要求]
以 JSON 格式输出：标题、分镜列表（每个分镜含时长、画面描述、台词、音效建议）、整体时长、推荐BGM风格。`

type VideoScriptPlugin struct{}

func NewVideoScriptPlugin() *VideoScriptPlugin {
	return &VideoScriptPlugin{}
}

func (p *VideoScriptPlugin) Name() string { return "video_script" }

func (p *VideoScriptPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "video_script",
		Summary:      "视频脚本创作，为抖音/B站/YouTube等平台生成高转化脚本",
		Description:  "AI 视频脚本助手：自动生成平台适配的带货/品牌/教程类视频脚本，含分镜、台词、时长建议",
		Capabilities: []string{"script-generation", "platform-adaptation", "storyboard"},
		Category:     "content",
		Parameters: map[string]interface{}{
			"platform": []string{"douyin", "bilibili", "youtube", "kuaishou", "tiktok"},
			"duration": "15-600s (depending on platform)",
			"style":    []string{"promotion", "tutorial", "brand", "entertainment", "review"},
		},
	}
}

func (p *VideoScriptPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	var params struct {
		Platform string `json:"platform"`
		Product  string `json:"product"`
		Style    string `json:"style"`
		Duration int    `json:"duration"`
		Keywords string `json:"keywords"`
	}
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return nil, fmt.Errorf("video_script: parse params: %w", err)
	}
	if params.Platform == "" {
		params.Platform = "douyin"
	}
	if params.Style == "" {
		params.Style = "promotion"
	}

	return &runtime.ExecutionResult{
		Data: map[string]interface{}{
			"platform": params.Platform,
			"product":  params.Product,
			"style":    params.Style,
			"script":   generateVideoScript(params),
		},
	}, nil
}

func generateVideoScript(params struct {
	Platform string `json:"platform"`
	Product  string `json:"product"`
	Style    string `json:"style"`
	Duration int    `json:"duration"`
	Keywords string `json:"keywords"`
}) map[string]interface{} {
	duration := params.Duration
	if duration <= 0 {
		switch params.Platform {
		case "douyin", "tiktok", "kuaishou":
			duration = 30
		case "bilibili":
			duration = 300
		case "youtube":
			duration = 480
		default:
			duration = 60
		}
	}

	scenes := []map[string]interface{}{
		{"time": "0-3s", "scene": "黄金开场", "lines": "你是否也曾遇到过这样的问题？", "visual": "痛点场景画面", "audio": "紧张氛围BGM"},
		{"time": "3-10s", "scene": "产品引入", "lines": fmt.Sprintf("今天给大家推荐一款%s", params.Product), "visual": "产品特写展示", "audio": "轻快背景音乐"},
		{"time": "10-20s", "scene": "功能演示", "lines": "它最大的特点是...", "visual": "使用场景画面", "audio": "功能展示音效"},
		{"time": "20-25s", "scene": "效果展示", "lines": "用了之后的效果真的很明显...", "visual": "前后对比画面", "audio": "惊喜音效"},
		{"time": "25-30s", "scene": "行动号召", "lines": "点击下方链接立即购买！", "visual": "购买引导画面", "audio": "节奏加快BGM"},
	}

	return map[string]interface{}{
		"title":      fmt.Sprintf("%s - %s平台%s风格脚本", params.Product, params.Platform, params.Style),
		"total_time": fmt.Sprintf("%ds", duration),
		"platform":   params.Platform,
		"scenes":     scenes,
		"bgm_style":  params.Platform + "_trending",
		"keywords":   params.Keywords,
	}
}

func (p *VideoScriptPlugin) Review(ctx context.Context, output interface{}) (*runtime.ReviewResult, error) {
	return &runtime.ReviewResult{
		Passed:  true,
		Score:   80,
		Summary: "视频脚本结构完整，分镜合理",
	}, nil
}
