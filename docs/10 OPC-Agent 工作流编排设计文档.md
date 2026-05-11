# OPC-Agent 工作流编排设计文档

> **版本**：v1.0  
> **日期**：2026-05-11  
> **关联文档**：[需求说明书](./01%20多任务智能协助系统（OPC-Agent）——项目需求说明书.md)  
> **项目代号**：OPC-Agent

---

## 1. 设计目标

将多个 Agent 串联或并联组成自动化工作流，让 OPC 创业者用"搭积木"的方式构建业务处理流水线。

### 典型场景

| 场景 | 流程 |
|------|------|
| **客户投诉自动处理** | 邮件分类 → 情感分析 → 回复建议生成 → 工单创建 |
| **竞品周报自动生成** | 竞品信息采集(并行) → 数据汇总 → SWOT 分析 → 报告撰写 |
| **内容多平台发布** | 文案生成 → 小红书适配(并行) → 微博适配(并行) → 发布确认 |
| **会议跟进闭环** | 会议纪要 → 待办提取 → 邮件通知各负责人 |

---

## 2. 工作流定义格式

### 2.1 YAML 工作流定义

```yaml
# workflows/customer_complaint.yaml
name: 客户投诉自动处理
description: 自动分类投诉邮件并生成回复
version: "1.0"
trigger: manual  # manual | schedule | webhook

steps:
  - id: classify
    agent: email_sorter
    input: "${trigger.input}"
    output: classify_result

  - id: analyze_sentiment
    agent: sentiment_analyzer
    input: "${classify_result.output.reply_suggestion}"
    output: sentiment_result

  - id: generate_reply
    agent: copywriter
    input: |
      投诉内容：${trigger.input}
      分类结果：${classify_result.output.category}
      情感分析：${sentiment_result.output.score}
      请生成一封专业、诚恳的回复邮件。
    output: final_reply

  - id: notify
    agent: email_sender
    input: "${final_reply.output}"
    output: send_result
    condition: "${classify_result.output.urgency} == '高'"
```

### 2.2 并行工作流

```yaml
# workflows/weekly_report.yaml
name: 竞品周报生成
trigger: schedule
schedule: "0 9 * * 1"  # 每周一 9:00

steps:
  - id: collect_competitors
    parallel:
      - agent: web_scraper
        input: "竞争对手A最新动态"
        output: comp_a_data
      - agent: web_scraper
        input: "竞争对手B最新动态"
        output: comp_b_data
      - agent: web_scraper
        input: "行业新闻"
        output: industry_news
    aggregate: combine  # 等待所有并行完成，合并结果

  - id: analyze
    agent: competitive_analysis
    input: |
      竞品A: ${comp_a_data.output}
      竞品B: ${comp_b_data.output}
      行业: ${industry_news.output}
    output: analysis_result

  - id: write_report
    agent: copywriter
    input: "${analysis_result.output}，请整理为周报格式"
    output: final_report
```

---

## 3. 核心数据结构

```go
// Workflow 工作流定义
type Workflow struct {
    Name        string       `yaml:"name"`
    Description string       `yaml:"description"`
    Version     string       `yaml:"version"`
    Trigger     string       `yaml:"trigger"`     // manual | schedule | webhook
    Schedule    string       `yaml:"schedule,omitempty"` // cron 表达式
    Steps       []Step       `yaml:"steps"`
    Variables   []Variable   `yaml:"variables,omitempty"`
}

// Step 单个步骤（串行）或并行组
type Step struct {
    ID        string        `yaml:"id"`
    Agent     string        `yaml:"agent,omitempty"`   // 串行：指定 Agent
    Input     string        `yaml:"input"`             // 支持 ${} 变量引用
    Output    string        `yaml:"output"`            // 输出变量名
    Parallel  []ParallelStep `yaml:"parallel,omitempty"` // 并行子步骤
    Aggregate string        `yaml:"aggregate,omitempty"` // combine | first
    Condition string        `yaml:"condition,omitempty"` // 条件执行
    Retry     *RetryConfig  `yaml:"retry,omitempty"`
}

// ParallelStep 并行子步骤
type ParallelStep struct {
    ID     string `yaml:"id"`
    Agent  string `yaml:"agent"`
    Input  string `yaml:"input"`
    Output string `yaml:"output"`
}

// RetryConfig 重试配置
type RetryConfig struct {
    MaxAttempts int `yaml:"max_attempts"`
    DelayMs     int `yaml:"delay_ms"`
}

// WorkflowEngine 工作流引擎
type WorkflowEngine struct {
    loader    *runtime.Loader
    client    runtime.ModelClient
    sessions  map[string]*WorkflowSession
}

// WorkflowSession 工作流执行会话
type WorkflowSession struct {
    ID        string
    Workflow  *Workflow
    Status    string    // running | paused | completed | failed
    StartTime time.Time
    Context   map[string]interface{} // 步骤间的变量传递
    Logs      []StepLog
}

type StepLog struct {
    StepID    string
    Status    string
    StartTime time.Time
    Duration  time.Duration
    Input     string
    Output    string
    Error     string
}
```

---

## 4. 变量引用系统

使用 `${}` 语法在工作流步骤间传递数据：

```yaml
# 变量引用规则
${trigger.input}              # 工作流触发时的原始输入
${classify_result.output}     # 引用上一步的完整输出
${classify_result.output.category}  # 引用输出中的特定字段
${analyze_result.output.score}      # 嵌套字段访问

# 内置变量
${workflow.name}              # 当前工作流名称
${workflow.started_at}        # 开始时间
${env.API_KEY}                # 环境变量
```

### 变量解析器

```go
// ResolveVariables 递归解析 ${} 变量引用
func (e *WorkflowEngine) ResolveVariables(ctx context.Context, 
    template string, ctxMap map[string]interface{}) (string, error) {
    re := regexp.MustCompile(`\$\{([^}]+)\}`)
    return re.ReplaceAllStringFunc(template, func(match string) string {
        path := match[2 : len(match)-1]
        value, err := resolvePath(ctxMap, path)
        if err != nil {
            return match // 保持原样
        }
        return fmt.Sprintf("%v", value)
    }), nil
}
```

---

## 5. 工作流引擎实现

### 5.1 串行执行

```
输入 → Step1 → Step2 → Step3 → 输出
         │        │        │
         ▼        ▼        ▼
    变量写入    变量写入    变量写入
    ctxMap     ctxMap     ctxMap
```

```go
func (e *WorkflowEngine) ExecuteSerial(ctx context.Context, 
    step Step, ctxMap map[string]interface{}) (interface{}, error) {
    
    // 解析条件
    if step.Condition != "" {
        passed, _ := evalCondition(step.Condition, ctxMap)
        if !passed {
            return nil, nil // 跳过此步骤
        }
    }
    
    // 解析输入变量
    input, _ := e.ResolveVariables(ctx, step.Input, ctxMap)
    
    // 获取 Agent 并执行
    plugin, _ := e.loader.Get(step.Agent)
    result, err := plugin.Execute(ctx, input, map[string]interface{}{
        "model_client": e.client,
    })
    if err != nil && step.Retry != nil {
        // 重试逻辑
        for i := 0; i < step.Retry.MaxAttempts; i++ {
            time.Sleep(time.Duration(step.Retry.DelayMs) * time.Millisecond)
            result, err = plugin.Execute(ctx, input, map[string]interface{}{
                "model_client": e.client,
            })
            if err == nil { break }
        }
    }
    
    // 保存输出到上下文
    ctxMap[step.Output] = result.Data
    return result.Data, err
}
```

### 5.2 并行执行

```
         ┌─ Step A ─┐
输入 ────┤─ Step B ─├──── 聚合 → 输出
         └─ Step C ─┘
```

```go
func (e *WorkflowEngine) ExecuteParallel(ctx context.Context, 
    steps []ParallelStep, ctxMap map[string]interface{}) (interface{}, error) {
    
    var wg sync.WaitGroup
    results := make(map[string]interface{})
    errCh := make(chan error, len(steps))
    
    for _, ps := range steps {
        wg.Add(1)
        go func(s ParallelStep) {
            defer wg.Done()
            input, _ := e.ResolveVariables(ctx, s.Input, ctxMap)
            plugin, _ := e.loader.Get(s.Agent)
            result, err := plugin.Execute(ctx, input, map[string]interface{}{
                "model_client": e.client,
            })
            if err != nil {
                errCh <- err
                return
            }
            ctxMap[s.Output] = result.Data
            results[s.ID] = result.Data
        }(ps)
    }
    
    wg.Wait()
    close(errCh)
    
    for err := range errCh {
        if err != nil { return nil, err }
    }
    
    // 聚合策略
    return results, nil // combine: 返回所有结果
}
```

---

## 6. CLI 交互

### 新增命令

```bash
# 列出可用工作流
> workflows
📋 可用工作流（3）：
  customer_complaint  客户投诉自动处理    manual
  weekly_report       竞品周报生成        schedule (每周一 9:00)
  content_publish     多平台内容发布      manual

# 执行工作流
> run customer_complaint 处理这封投诉邮件：产品有质量问题要求退款

# 查看工作流状态
> workflow status customer_complaint
📊 状态：running（步骤 2/4）
  ✅ classify     邮件分类        1.2s
  🔄 analyze      情感分析        进行中...
  ⏳ generate     回复生成        等待中
  ⏳ notify       通知发送        等待中

# 查看工作流历史
> workflow history
  2026-05-11 10:30  customer_complaint  ✅ 完成  12.5s
  2026-05-11 09:15  weekly_report       ✅ 完成  45.2s
  2026-05-10 14:00  content_publish     ❌ 失败  步骤3超时
```

### 工作流 DSL 简写

在对话中直接定义临时工作流：

```
> @classify 投诉邮件 | @analyze 情感分析 | @reply 生成回复
```

等价于三步串行工作流，自动创建临时执行。

---

## 7. Web UI 工作流设计器

### 7.1 工作流列表页

```
┌────────────────────────────────────────────────────┐
│ 📋 工作流                           [+ 创建工作流] │
├────────────────────────────────────────────────────┤
│                                                    │
│ ┌─ 客户投诉自动处理 ──────────────────────────┐    │
│ │  触发: manual · 4 步 · 含条件分支            │    │
│ │  最后运行: 5m ago ✅                        │    │
│ │  [▶ 运行] [✏️ 编辑] [📊 历史]              │    │
│ └──────────────────────────────────────────────┘    │
│                                                    │
│ ┌─ 竞品周报生成 ─────────────────────────────┐    │
│ │  触发: schedule (每周一 9:00) · 3 步 · 并行  │    │
│ │  最后运行: 昨天 09:02 ✅                    │    │
│ │  [▶ 运行] [✏️ 编辑] [📊 历史]              │    │
│ └──────────────────────────────────────────────┘    │
│                                                    │
│ ┌─ 多平台内容发布 ───────────────────────────┐    │
│ │  触发: manual · 1+3 步 · 并行分发           │    │
│ │  最后运行: 2天前 ❌ (步骤3超时)              │    │
│ │  [▶ 运行] [✏️ 编辑] [📊 历史]              │    │
│ └──────────────────────────────────────────────┘    │
└────────────────────────────────────────────────────┘
```

### 7.2 工作流可视化编辑器

```
┌────────────────────────────────────────────────────┐
│ ✏️ 编辑: 客户投诉自动处理              [💾 保存]    │
├────────────────────────────────────────────────────┤
│                                                    │
│  [触发] → [📧 邮件分类] → [💬 情感分析] → [✍️ 回复] → [📨 发送] │
│  manual     email_sorter  sentiment    copywriter  email_sender │
│               │              │                          │
│               ▼              ▼                          │
│           输出:classify   输出:sentiment           条件:urgency=高  │
│                                                    │
│  ┌─ 步骤配置 ─────────────────────────────┐         │
│  │  ID: classify                            │         │
│  │  Agent: email_sorter                     │         │
│  │  输入: ${trigger.input}                   │         │
│  │  输出变量: classify_result               │         │
│  │  ┌─ 条件分支 ──────┐                     │         │
│  │  │ urgency="高" → 发送通知               │         │
│  │  │ urgency="低" → 仅记录                 │         │
│  │  └──────────────────┘                     │         │
│  └──────────────────────────────────────────┘         │
└────────────────────────────────────────────────────┘
```

---

## 8. 实现路线

| 阶段 | 内容 | 工时 |
|------|------|------|
| **Phase 1** | 工作流 YAML 解析 + 串行执行引擎 | 2d |
| **Phase 2** | 并行执行 + 变量引用系统 | 1.5d |
| **Phase 3** | CLI 命令（workflows / run / status） | 1d |
| **Phase 4** | 条件分支 + 重试策略 | 1d |
| **Phase 5** | Web UI 工作流设计器 | 2d |

---

*文档结束*
