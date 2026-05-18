package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

var batchSystemPrompt = `你是一个批量处理协调专家。你的任务是将批量输入拆分为独立的小任务，合并相似结果。

[角色边界]
- 只做任务拆分和结果合并，不做单个任务的具体执行
- 保持每个子任务的独立性，无交叉依赖
- 输出必须严格遵循 JSON 格式

[输入格式]
- tasks: 需要批量处理的任务列表（每个任务有 id 和 content）
- batch_size: 每批处理数量（默认 5）
- strategy: 处理策略（sequential 顺序, parallel_merge 并行合并）

[输出要求]
- plan: 批处理执行计划
  - batches: 批次数组，每批含 task_ids
  - strategy: 使用的策略
  - estimated_rounds: 预估轮次
- summary: 批次统计摘要`

type BatchPlugin struct{}

func (p *BatchPlugin) Name() string { return "batch" }

func (p *BatchPlugin) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "batch",
		Summary:      "批量任务拆分与结果合并协调器",
		Version:      "1.0.0",
		Tags:         []string{"batch", "processing", "pipeline"},
		RequiresHITL: true,
		ModelName:    "gemini-2.0-flash",
	}
}

func (p *BatchPlugin) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("batch: 未配置模型 API")
	}
	modelName := resolveModelName(opts, p.Info().ModelName)

	var req struct {
		Tasks     []map[string]string `json:"tasks"`
		BatchSize int                 `json:"batch_size"`
		Strategy  string              `json:"strategy"`
	}
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		return nil, fmt.Errorf("batch: invalid input: %w", err)
	}
	if len(req.Tasks) == 0 {
		return nil, fmt.Errorf("batch: tasks is empty")
	}
	if req.BatchSize <= 0 {
		req.BatchSize = 5
	}
	if req.Strategy == "" {
		req.Strategy = "sequential"
	}

	var sb strings.Builder
	sb.WriteString(batchSystemPrompt)
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("总任务数: %d\n", len(req.Tasks)))
	sb.WriteString(fmt.Sprintf("每批大小: %d\n", req.BatchSize))
	sb.WriteString(fmt.Sprintf("处理策略: %s\n", req.Strategy))
	sb.WriteString("\n任务列表:\n")
	for _, t := range req.Tasks {
		id := t["id"]
		content := t["content"]
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		sb.WriteString(fmt.Sprintf("- [%s] %s\n", id, content))
	}

	modelResp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: "你是一个批量处理专家。严格按照 system prompt 的格式输出 JSON。",
		UserMessage:  sb.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("batch: model call: %w", err)
	}

	result := map[string]interface{}{
		"total_tasks": len(req.Tasks),
		"batch_size":  req.BatchSize,
		"strategy":    req.Strategy,
		"plan":        modelResp.Content,
	}

	return &runtime.ExecutionResult{
		Data: result,
		TokenUsage: runtime.TokenUsage{
			InputTokens:  modelResp.InputTokens,
			OutputTokens: modelResp.OutputTokens,
			ModelName:    modelName,
		},
	}, nil
}

func (p *BatchPlugin) Review(ctx context.Context, output interface{}) (*runtime.ReviewResult, error) {
	return &runtime.ReviewResult{Passed: true, Score: 1.0, Summary: "batch plan generated"}, nil
}
