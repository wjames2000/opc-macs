package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type MeetingMinutesPlugin struct{}

func (p *MeetingMinutesPlugin) Name() string { return "meeting_minutes" }

func (p *MeetingMinutesPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:    "meeting_minutes",
		Summary: "将会议讨论整理为结构化会议纪要",
		Version: "1.0.0",
		Tags:    []string{"meeting", "minutes", "productivity", "document"},
	}
}

var meetingMinutesSystemPrompt = `你是专业的会议纪要撰写助手。请将会议讨论内容整理为结构化的会议纪要。

[输出格式 JSON]
{
  "title": "会议主题",
  "time": "会议时间",
  "participants": ["参会人1", "参会人2"],
  "agenda": ["议题1", "议题2"],
  "decisions": ["决策1", "决策2"],
  "action_items": [
    {"task": "待办事项", "owner": "负责人", "deadline": "截止日期"}
  ],
  "next_steps": "下一步计划"
}`

func (p *MeetingMinutesPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("meeting_minutes: 未配置模型 API")
	}
	modelName := resolveModelName(opts)

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: meetingMinutesSystemPrompt,
		UserMessage:  input,
	})
	if err != nil {
		return nil, fmt.Errorf("meeting_minutes: 模型调用失败：%w", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		return nil, fmt.Errorf("meeting_minutes: 输出解析失败：%w", err)
	}

	return &runtime.ExecutionResult{
		Data:       output,
		TokenUsage: runtime.TokenUsage{InputTokens: resp.InputTokens, OutputTokens: resp.OutputTokens, ModelName: modelName},
		RawTrace:   resp.RawResponse,
	}, nil
}

func (p *MeetingMinutesPlugin) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	return &runtime.ReviewResult{Passed: true, Score: 5.0, Summary: "meeting_minutes: 完成"}, nil
}

func resolveModelName(opts map[string]interface{}) string {
	if m, ok := opts["model_name"].(string); ok && m != "" {
		return m
	}
	return "deepseek-v4-flash"
}

var Agent MeetingMinutesPlugin
