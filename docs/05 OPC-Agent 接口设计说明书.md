# OPC-Agent 接口设计说明书

> **版本**：v1.0  
> **日期**：2026-05-09  
> **关联文档**：[项目需求说明书 v1.1](./01%20多任务智能协助系统（OPC-Agent）——项目需求说明书.md)，[概要设计说明书](./03%20OPC-Agent%20概要设计说明书.md)  
> **项目代号**：OPC-Agent

---

## 1. 接口概览

| 接口 | 位置 | 用途 | 调用方 |
|------|------|------|--------|
| **AgentPlugin** | `internal/runtime/plugin.go` | 所有 Agent 插件必须实现的统一接口 | Plugin Runtime 加载器 |
| **MemoryStore** | `internal/memory/store.go` | 记忆存储与检索 | 各 Agent 插件 |
| **HITL 接口** | `internal/hitl/handler.go` | 人工确认交互 | Harness |
| **Reviewer 接口** | `internal/agent/reviewer.go` | 输出审查 | Harness |
| **Config 接口** | `internal/config/config.go` | 配置加载 | 全局 |

---

## 2. AgentPlugin 接口设计

### 2.1 AgentPlugin 接口定义

```go
package runtime

import "context"

// PluginInfo 插件元信息（L1 发现层）
type PluginInfo struct {
    Name        string   // Agent 名称，必须唯一
    Summary     string   // 简述（≤30 字）
    Version     string   // 插件版本
    Tags        []string // 标签协助路由匹配
    RequiresHITL bool    // 是否需要人工确认
}

// ExecutionResult 执行结果
type ExecutionResult struct {
    Data       interface{} // 结构化输出（由产物契约定义）
    TokenUsage TokenUsage  // token 消耗统计
}

// TokenUsage token 消耗统计
type TokenUsage struct {
    InputTokens  int
    OutputTokens int
}

// AgentPlugin 是所有 Agent 插件必须实现的接口
// 每个插件在 plugin.go 中定义导出符号 var Agent AgentPlugin
type AgentPlugin interface {
    // Name 返回 Agent 唯一名称
    Name() string

    // Info 返回插件元信息（用于 L1 发现）
    Info() PluginInfo

    // Execute 执行 Agent 核心逻辑
    // input: 用户输入
    // opts: 扩展参数（记忆上下文、配置等）
    Execute(ctx context.Context, input string, opts map[string]interface{}) (*ExecutionResult, error)

    // Review 可选：对自身输出的审查（未实现时由主程序的 Reviewer Agent 完成）
    Review(ctx context.Context, output interface{}) (*ReviewResult, error)
}
```

### 2.2 Plugin Loader 实现

```go
package runtime

import (
    "fmt"
    "os"
    "path/filepath"
    "plugin"
)

// PluginSymbol 是插件入口的导出符号名
const PluginSymbol = "Agent"

// Loader 插件加载器
type Loader struct {
    pluginsDir string
    plugins    map[string]AgentPlugin
}

func NewLoader(pluginsDir string) *Loader {
    return &Loader{
        pluginsDir: pluginsDir,
        plugins:    make(map[string]AgentPlugin),
    }
}

// LoadAll 扫描 pluginsDir 下的所有 .so 文件并加载
func (l *Loader) LoadAll() error {
    entries, err := os.ReadDir(l.pluginsDir)
    if err != nil {
        return fmt.Errorf("runtime: 读取插件目录 %s 失败：%w", l.pluginsDir, err)
    }

    for _, entry := range entries {
        if entry.IsDir() || filepath.Ext(entry.Name()) != ".so" {
            continue
        }
        if err := l.loadFile(filepath.Join(l.pluginsDir, entry.Name())); err != nil {
            fmt.Printf("[Warning] 加载插件 %s 失败：%v\n", entry.Name(), err)
        }
    }

    return nil
}

// loadFile 加载单个 .so 文件
func (l *Loader) loadFile(soPath string) error {
    p, err := plugin.Open(soPath)
    if err != nil {
        return fmt.Errorf("plugin.Open 失败：%w", err)
    }

    sym, err := p.Lookup(PluginSymbol)
    if err != nil {
        return fmt.Errorf("Lookup(%s) 失败：%w", PluginSymbol, err)
    }

    agent, ok := sym.(AgentPlugin)
    if !ok {
        return fmt.Errorf("插件 %s 未实现 AgentPlugin 接口", soPath)
    }

    info := agent.Info()
    if _, exists := l.plugins[info.Name]; exists {
        return fmt.Errorf("Agent '%s' 已存在，请检查插件目录", info.Name)
    }

    l.plugins[info.Name] = agent
    fmt.Printf("[Plugin] 已加载：%s v%s\n", info.Name, info.Version)
    return nil
}

// Get 按名称获取已加载的插件
func (l *Loader) Get(name string) (AgentPlugin, bool) {
    p, ok := l.plugins[name]
    return p, ok
}

// List 返回所有已加载插件的 L1 元信息
func (l *Loader) List() []PluginInfo {
    result := make([]PluginInfo, 0, len(l.plugins))
    for _, p := range l.plugins {
        result = append(result, p.Info())
    }
    return result
}

// Count 返回已加载插件数量
func (l *Loader) Count() int {
    return len(l.plugins)
}
```

### 2.3 Agent 插件编写示例

```go
// plugins/copywriter/plugin.go
package main

import (
    "context"
    "fmt"
    "github.com/yourname/opc-agent/internal/runtime"
)

// CopywriterPlugin 文案生成 Agent 插件
type CopywriterPlugin struct{}

func (p *CopywriterPlugin) Name() string { return "copywriter" }

func (p *CopywriterPlugin) Info() runtime.PluginInfo {
    return runtime.PluginInfo{
        Name:        "copywriter",
        Summary:     "根据产品信息生成三段式营销文案",
        Version:     "1.0.0",
        Tags:        []string{"marketing", "copywriting"},
        RequiresHITL: true,
    }
}

func (p *CopywriterPlugin) Execute(ctx context.Context,
    input string, opts map[string]interface{}) (*runtime.ExecutionResult, error) {

    // 1. 从 opts 中提取注入的记忆上下文
    memories, _ := opts["memories"].([]string)

    // 2. 使用 ADK 构建 Agent 并执行
    //    （每个插件可独立使用 ADK 框架）
    agent := buildCopywriterAgent(memories)

    // 3. 执行并返回结构化结果
    result, err := agent.Execute(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("copywriter plugin: %w", err)
    }

    return &runtime.ExecutionResult{
        Data: result,
        TokenUsage: runtime.TokenUsage{
            InputTokens:  result.InputTokens,
            OutputTokens: result.OutputTokens,
        },
    }, nil
}

func (p *CopywriterPlugin) Review(ctx context.Context,
    output interface{}) (*runtime.ReviewResult, error) {
    // 可选实现：插件的自审查逻辑
    // 返回 nil 表示由主程序的 Reviewer Agent 处理
    return nil, nil
}

// 插件入口符号 —— 必须导出
var Agent CopywriterPlugin
```

### 2.4 插件构建

```bash
# 编译为 .so 共享库
cd plugins/copywriter
go build -buildmode=plugin -o ../../build/plugins/copywriter.so .
```

### 2.5 Router Agent（内置在主程序中）

```go
// Router Agent 不作为插件，内置于主程序
// 因为它需要直接管理插件加载器的 L1 元数据

// NewRouterAgent 创建路由 Agent
func NewRouterAgent(loader *runtime.Loader) *adk.Agent {
    l1Context := buildL1Context(loader.List())

    agent := &adk.Agent{
        Name:        "router",
        Instruction: l1Context,
        Model:       cfg.Model.Name,
    }
    return agent
}

// buildL1Context 构建 L1 上下文：仅含插件名称和简述
func buildL1Context(plugins []runtime.PluginInfo) string {
    var sb strings.Builder
    sb.WriteString("你是 OPC-Agent 的路由智能体。根据用户输入判断意图，分发到对应 Agent。\n\n")
    sb.WriteString("可用 Agent 插件：\n")
    for _, p := range plugins {
        sb.WriteString(fmt.Sprintf("- %s：%s\n", p.Name, p.Summary))
    }
    sb.WriteString("\n如果无法匹配任何技能，返回 \"unknown\" 并提示用户重新描述。")
    return sb.String()
}
```

---

## 3. Skill 接口设计（Plugin 内部使用）

> Skill 不再作为主程序的独立模块存在。每个 Agent 插件在自己的 `SKILL.md` 中定义 Skill 内容，通过 `AgentPlugin.Info()` 暴露 L1 元数据，通过 `AgentPlugin.Execute()` 内部加载 L2/L3 指令。以下接口供插件开发者参考。

### 3.1 Skill 定义（插件内部）

```go
// Meta 技能元数据（L1 发现层）
type Meta struct {
    Name        string   `yaml:"name"`        // 技能名称
    Summary     string   `yaml:"summary"`      // 简述（≤30 字）
    Version     string   `yaml:"version"`
    Tags        []string `yaml:"tags"`         // 标签协助路由匹配
    RequiresHITL bool   `yaml:"requires_hitl"` // 是否需要人工确认
}

// Skill 完整技能定义
type Skill struct {
    Meta
    Instruction string           // L2 完整指令（来自 SKILL.md）
    Contract    *ArtifactContract // 产物契约定义
    Checkpoints []Checkpoint     // Reviewer 检查清单
    Resources   []string         // L3 参考资源文件路径
}
```

### 3.2 SKILL.md 契约格式

每个 Skill 目录下必须包含 `SKILL.md`，格式如下：

```markdown
# Skill: [技能名称]

## 元数据（L1）
- 名称：copywriting
- 简述：根据产品信息生成三段式营销文案
- 标签：marketing, copywriting, social-media
- 需 HITL：是（发布前需确认）

## 指令（L2）
[Agent 角色定义]
你是专业的营销文案撰写专家。

[执行步骤]
1. 理解产品名称、核心卖点和目标风格
2. 生成短文案（≤20 字，适合标题）
3. 生成长文案（100-200 字，适合详情页）
4. 生成社交媒体文案（≤80 字，适合朋友圈/推文）

[输出格式]
{
  "short_copy": "string",
  "long_copy": "string",
  "social_copy": "string",
  "style": "string"
}

[角色边界]
- 只做文案生成，不做市场分析
- 不使用夸张/虚假宣传词汇
- 不使用禁止词列表中的词汇

## 检查清单（Reviewer 使用）
- [ ] 短文案长度 ≤ 20 字
- [ ] 长文案 100-200 字
- [ ] 社交文案 ≤ 80 字
- [ ] 无禁止词
- [ ] 风格与要求一致
- [ ] JSON 格式正确

## 禁止词
- "guaranteed"
- "100% safe"
- "best in class"（未经证实）
```

---

## 4. Memory Store 接口

```go
package memory

import "context"

// MemoryEntry 记忆条目
type MemoryEntry struct {
    ID           string            `json:"id"`
    CreatedAt    time.Time         `json:"created_at"`
    TaskType     string            `json:"task_type"`
    UserInput    string            `json:"user_input"`
    AgentOutput  string            `json:"agent_output"`
    KeyDecisions []string          `json:"key_decisions"`
    Embedding    []float32         `json:"embedding"`
    Metadata     map[string]string `json:"metadata"`
}

// MemoryStore 记忆存储接口
type MemoryStore interface {
    Store(ctx context.Context, entry *MemoryEntry) error
    Search(ctx context.Context, query []float32, topK int) ([]*MemoryEntry, error)
    Recall(ctx context.Context, text string, topK int) ([]*MemoryEntry, error)
    Count(ctx context.Context) (int, error)
    Close() error
}

// Recall 快速使用示例：
// 1. 调用方传入用户输入文本
// 2. Recall 内部调用 LLM embedding API 生成向量
// 3. 执行 Search 返回 TopK 结果
```

### 记忆检索在 Agent 中的使用

```go
// memoryRetrievalTool 创建 ADK 工具：检索记忆注入上下文
func memoryRetrievalTool(store memory.MemoryStore) adk.Tool {
    return adk.Tool{
        Name:        "retrieve_memory",
        Description: "检索与当前任务相似的历史记忆，用于参考",
        Callback: func(ctx context.Context, input string) (string, error) {
            entries, err := store.Recall(ctx, input, 5)
            if err != nil || len(entries) == 0 {
                return "未找到相关历史记忆。", nil
            }
            var sb strings.Builder
            sb.WriteString("以下是与当前任务相关的历史记忆（供参考）：\n")
            for i, e := range entries {
                sb.WriteString(fmt.Sprintf("\n--- 记忆 %d ---\n", i+1))
                sb.WriteString(fmt.Sprintf("任务类型：%s\n", e.TaskType))
                sb.WriteString(fmt.Sprintf("用户输入：%s\n", e.UserInput))
                sb.WriteString(fmt.Sprintf("输出摘要：%s\n", e.AgentOutput))
            }
            sb.WriteString("\n[注意] 以上信息来源于历史记忆，仅供参考，请根据当前实际需求处理。")
            return sb.String(), nil
        },
    }
}
```

---

## 5. HITL 接口设计

```go
package hitl

// Operation 敏感操作定义
type Operation struct {
    Type        string // 操作类型（send_email / publish_post）
    Description string // 操作描述，供用户决策
    Payload     interface{} // 操作数据
}

// Handler HITL 处理器
type Handler struct {
    reader io.Reader // 可替换的输入源（终端 / Web UI）
    writer io.Writer
}

func NewHandler(reader io.Reader, writer io.Writer) *Handler

// Confirm 阻塞等待用户确认
// 返回 true = 用户同意（y/yes），false = 拒绝（n/no）
func (h *Handler) Confirm(ctx context.Context, op Operation) (bool, error) {
    fmt.Fprintf(h.writer, "\n⚠️  敏感操作需要确认：\n")
    fmt.Fprintf(h.writer, "操作类型：%s\n", op.Type)
    fmt.Fprintf(h.writer, "操作描述：%s\n", op.Description)
    fmt.Fprintf(h.writer, "请输入 y/n 确认：")

    var input string
    _, err := fmt.Fscanln(h.reader, &input)
    if err != nil {
        return false, fmt.Errorf("读取输入失败：%w", err)
    }

    input = strings.TrimSpace(strings.ToLower(input))
    return input == "y" || input == "yes", nil
}
```

### 集成到 Harness

```go
// Harness 中的 HITL 检查点
func (h *Harness) checkHITL(ctx context.Context, op hitl.Operation) error {
    if !h.shouldBlock(op.Type) {
        return nil // 不在敏感操作列表中，放行
    }

    approved, err := h.hitlHandler.Confirm(ctx, op)
    if err != nil {
        return fmt.Errorf("HITL 错误：%w", err)
    }
    if !approved {
        return fmt.Errorf("用户拒绝操作：%s", op.Type)
    }
    return nil
}
```

---

## 6. Reviewer 接口设计

```go
package agent

// ReviewResult 审查结果
type ReviewResult struct {
    Passed      bool               // 是否通过
    Score       float32            // 评分 0.0-5.0
    CheckResults []CheckResult     // 逐项检查结果
    Summary     string             // 审查摘要
    ShouldRetry bool               // 是否建议重试
}

// CheckResult 单项检查结果
type CheckResult struct {
    Item    string // 检查项描述
    Passed  bool   // 是否通过
    Detail  string // 详情
}

// ReviewOutput 审查 Agent 的输出
// checkpoints 由对应插件的 SKILL.md 定义，通过插件 Info() 或其他方式获取
func ReviewOutput(ctx context.Context, agentOutput string, checkpoints []string) (*ReviewResult, error) {
    prompt := buildReviewPrompt(agentOutput, checkpoints)
    // 调用 LLM 审查
    response := callLLM(ctx, prompt)
    // 解析响应为 ReviewResult
    return parseReviewResult(response)
}
```

---

## 7. Config 接口设计

```go
package config

import "os"

// Config 全局配置
type Config struct {
    App     AppConfig     `yaml:"app"`
    Model   ModelConfig   `yaml:"model"`
    Runtime RuntimeConfig `yaml:"runtime"`   // 插件运行时配置
    Memory  MemoryConfig  `yaml:"memory"`
    Harness HarnessConfig `yaml:"harness"`
}

type AppConfig struct {
    Name    string `yaml:"name"`
    Version string `yaml:"version"`
}

type ModelConfig struct {
    Provider    string  `yaml:"provider"`
    Name        string  `yaml:"name"`
    Temperature float32 `yaml:"temperature"`
    MaxTokens   int     `yaml:"max_tokens"`
}

// RuntimeConfig 插件运行时配置
type RuntimeConfig struct {
    PluginsDir string `yaml:"plugins_dir"` // .so 插件目录，默认 "./build/plugins"
}

type MemoryConfig struct {
    Engine             string  `yaml:"engine"`              // embedded | pgvector
    TopK               int     `yaml:"top_k"`
    SimilarityThreshold float32 `yaml:"similarity_threshold"`
    StorePath          string  `yaml:"store_path"`
    PGConnStr          string  `yaml:"pg_conn_str,omitempty"`
}

type HarnessConfig struct {
    MaxStepsPerAgent int      `yaml:"max_steps_per_agent"`
    HITLOperations   []string `yaml:"hitl_operations"`
    ForbiddenWords   []string `yaml:"forbidden_words"`
}

type AppConfig struct {
    Name    string `yaml:"name"`
    Version string `yaml:"version"`
}

type ModelConfig struct {
    Provider    string  `yaml:"provider"`
    Name        string  `yaml:"name"`
    Temperature float32 `yaml:"temperature"`
    MaxTokens   int     `yaml:"max_tokens"`
}

type SkillConfig struct {
    Name    string `yaml:"name"`
    Path    string `yaml:"path"`
    Enabled bool   `yaml:"enabled"`
}

type MemoryConfig struct {
    Engine             string  `yaml:"engine"`              // embedded | pgvector
    TopK               int     `yaml:"top_k"`
    SimilarityThreshold float32 `yaml:"similarity_threshold"`
    StorePath          string  `yaml:"store_path"`
    PGConnStr          string  `yaml:"pg_conn_str,omitempty"`
}

type HarnessConfig struct {
    MaxStepsPerAgent int      `yaml:"max_steps_per_agent"`
    HITLOperations   []string `yaml:"hitl_operations"`
    ForbiddenWords   []string `yaml:"forbidden_words"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }
    // 设置默认值
    if cfg.Runtime.PluginsDir == "" {
        cfg.Runtime.PluginsDir = "./build/plugins"
    }
    if cfg.Memory.TopK <= 0 {
        cfg.Memory.TopK = 5
    }
    if cfg.Memory.SimilarityThreshold <= 0 {
        cfg.Memory.SimilarityThreshold = 0.7
    }
    return &cfg, nil
}
```

---

## 8. 产出契约（JSON Schema）

### 8.1 文案生成输出

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CopywritingOutput",
  "type": "object",
  "required": ["short_copy", "long_copy", "social_copy", "style"],
  "properties": {
    "short_copy": {
      "type": "string",
      "description": "短文案（≤20字）",
      "maxLength": 20
    },
    "long_copy": {
      "type": "string",
      "description": "长文案（100-200字）",
      "minLength": 100,
      "maxLength": 200
    },
    "social_copy": {
      "type": "string",
      "description": "社交媒体文案（≤80字）",
      "maxLength": 80
    },
    "style": {
      "type": "string",
      "description": "文案风格",
      "enum": ["科技简约", "温暖亲和", "专业正式", "年轻活力"]
    }
  }
}
```

### 8.2 邮件分类输出

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "EmailClassificationOutput",
  "type": "object",
  "required": ["category", "reason", "reply_suggestion"],
  "properties": {
    "category": {
      "type": "string",
      "description": "邮件类别",
      "enum": ["咨询", "投诉", "合作", "垃圾"]
    },
    "reason": {
      "type": "string",
      "description": "分类理由（20-50字）"
    },
    "reply_suggestion": {
      "type": "string",
      "description": "回复建议草稿"
    },
    "urgency": {
      "type": "string",
      "description": "紧急程度",
      "enum": ["低", "中", "高"]
    }
  }
}
```

---

## 9. 接口版本与兼容性

| 接口 | 稳定性 | 变更影响范围 |
|------|--------|-------------|
| MemoryStore | MVP 迭代中可能变化 | 仅 Engine 实现类 |
| Skill Meta 结构 | 稳定 | 所有 Skill 定义 |
| Agent 注册方式 | 稳定（ADK 框架决定） | main.go |
| JSON Schema 契约 | 稳定 | Agent 输出 + Reviewer |
| Config 结构 | MVP 迭代中可能增加字段 | 向后兼容 |

---

*文档结束*
