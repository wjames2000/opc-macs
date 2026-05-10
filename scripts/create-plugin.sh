#!/bin/bash
# OPC-Agent 插件脚手架生成工具
# 用法: ./scripts/create-plugin.sh <插件名> <简述> [标签...]

set -e

if [ $# -lt 2 ]; then
    echo "用法: $0 <插件名> <简述> [标签...]"
    echo "示例: $0 seo_optimizer 'SEO 优化建议' seo marketing content"
    exit 1
fi

NAME=$1
SUMMARY=$2
shift 2
TAGS=("$@")

# 验证插件名
if ! [[ $NAME =~ ^[a-z][a-z0-9_]+$ ]]; then
    echo "错误: 插件名只能包含小写字母、数字和下划线，且以字母开头"
    exit 1
fi

MODULE="github.com/wjames2000/opc-macs"
PLUGIN_DIR="plugins/$NAME"
STRUCT_NAME="$(echo $NAME | sed -r 's/(^|_)([a-z])/\U\2/g')Plugin"

# 创建目录
mkdir -p "$PLUGIN_DIR"
echo "📁 创建 $PLUGIN_DIR/"

# go.mod
cat > "$PLUGIN_DIR/go.mod" <<EOF
module $MODULE/plugins/$NAME

go 1.26.2

require $MODULE v0.0.0

replace $MODULE => ../..
EOF
echo "  ├── go.mod"

# SKILL.md
cat > "$PLUGIN_DIR/SKILL.md" <<EOF
# Skill: $NAME

## 元数据（L1）
- 名称：$NAME
- 简述：$SUMMARY
- 标签：${TAGS[*]}
- 需 HITL：是（发布前需确认）

## 指令（L2）
你是 $SUMMARY 专家。

[执行步骤]
1. 理解用户输入
2. 分析关键信息
3. 生成结构化输出

[输出格式]
{
  "result": "string"
}

## 检查清单（Reviewer 使用）
- [ ] 输出格式正确
- [ ] 内容完整
- [ ] 无禁止词

## 禁止词
- 示例禁止词
EOF
echo "  ├── SKILL.md"

# plugin.go
# 构建标签列表
TAGS_GO=$(printf '"%s", ' "${TAGS[@]}")
TAGS_GO="[]string{${TAGS_GO%, }}"

cat > "$PLUGIN_DIR/plugin.go" <<EOF
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"$MODULE/internal/runtime"
)

type $STRUCT_NAME struct{}

func (p *$STRUCT_NAME) Name() string { return "$NAME" }

func (p *$STRUCT_NAME) Info() runtime.PluginInfo {
	return runtime.PluginInfo{
		Name:         "$NAME",
		Summary:      "$SUMMARY",
		Version:      "1.0.0",
		Tags:         $TAGS_GO,
		RequiresHITL: true,
	}
}

func (p *$STRUCT_NAME) Execute(ctx context.Context, input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {
	client, _ := opts["model_client"].(runtime.ModelClient)
	if client == nil {
		return nil, fmt.Errorf("$NAME: 未配置模型 API")
	}
	modelName := resolveModelName(opts)

	memories, _ := opts["memories"].([]string)
	history, _ := opts["history"].(string)

	// 构建 prompt - 从 SKILL.md 加载
	systemPrompt := fmt.Sprintf("你是 %s。请根据用户输入生成输出。", "$SUMMARY")

	payload := input
	if len(memories) > 0 || history != "" {
		payload = ""
		if len(memories) > 0 {
			payload += "\n参考信息：\n"
			for i, m := range memories {
				if i >= 3 { break }
				payload += fmt.Sprintf("- %s\n", m)
			}
		}
		if history != "" {
			payload += "\n" + history
		}
		payload += "\n\n当前任务：\n" + input
	}

	resp, err := client.Call(ctx, runtime.ModelRequest{
		Model:        modelName,
		SystemPrompt: systemPrompt,
		UserMessage:  payload,
	})
	if err != nil {
		return nil, fmt.Errorf("$NAME: 模型调用失败：%w", err)
	}

	var output map[string]interface{}
	if err := json.Unmarshal([]byte(resp.Content), &output); err != nil {
		output = map[string]interface{}{"result": resp.Content}
	}

	return &runtime.ExecutionResult{
		Data:       output,
		TokenUsage: runtime.TokenUsage{InputTokens: resp.InputTokens, OutputTokens: resp.OutputTokens, ModelName: modelName},
		RawTrace:   resp.RawResponse,
	}, nil
}

func (p *$STRUCT_NAME) Review(_ context.Context, output interface{}) (*runtime.ReviewResult, error) {
	return &runtime.ReviewResult{
		Passed: true, Score: 5.0, ShouldRetry: false,
		Summary: "$NAME: 完成",
	}, nil
}

// 辅助函数：从 opts 解析模型名
func resolveModelName(opts map[string]interface{}) string {
	if m, ok := opts["model_name"].(string); ok && m != "" {
		return m
	}
	return "deepseek-v4-flash"
}

var Agent $STRUCT_NAME
EOF
echo "  └── plugin.go"

echo ""
echo "✅ 插件 $NAME 已创建!"
echo ""
echo "下一步:"
echo "  1. 编辑 $PLUGIN_DIR/SKILL.md 完善技能定义"
echo "  2. 编辑 $PLUGIN_DIR/plugin.go 实现 Execute 逻辑"
echo "  3. 在 internal/plugins/plugins.go 的 RegisterAll() 中注册:"
echo "     &${STRUCT_NAME}{},"
echo "  4. 编译并运行: make run"
