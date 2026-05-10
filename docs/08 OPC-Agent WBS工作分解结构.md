# OPC-Agent WBS 工作分解结构（细化版）

> **版本**：v2.0  
> **日期**：2026-05-10  
> **关联文档**：[需求说明书 v1.1](./01%20多任务智能协助系统（OPC-Agent）——项目需求说明书.md)，[详细设计说明书](./07%20OPC-Agent%20详细设计说明书.md)  
> **项目代号**：OPC-Agent  
> **总工期**：14 天（MVP）  
> **总任务数**：134 项

---

## WBS 概览

```
OPC-Agent MVP (134 tasks, ~90h)
│
├── WBS 1  项目启动与环境搭建             第 1 天     7 tasks
├── WBS 2  internal/config 配置模块       第 1-2 天   9 tasks
├── WBS 3  internal/harness Harness 控制  第 2-4 天  17 tasks
├── WBS 4  internal/memory 记忆引擎       第 3-5 天  15 tasks
├── WBS 5  internal/hitl HITL 模块        第 4 天     5 tasks
├── WBS 6  internal/runtime 插件运行时    第 4-6 天  15 tasks
├── WBS 7  internal/agent 内置 Agent     第 6-8 天  20 tasks
├── WBS 8  plugins/copywriter 文案插件    第 8-10 天 12 tasks
├── WBS 9  plugins/email_sorter 邮件插件  第 9-11 天 12 tasks
├── WBS 10 plugins/xhs_poster 小红书插件  第 10-12 天 12 tasks
├── WBS 11 cmd/opc-agent 主程序           第 11-12 天 11 tasks
├── WBS 12 测试                           第 12-14 天 18 tasks
└── WBS 13 工程化与文档                   第 14 天     6 tasks
```

---

## WBS 1 — 项目启动与环境搭建

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 1.1 | 初始化 Go module：`go mod init github.com/yourname/opc-agent` | 15 | — | go.mod | `go build ./...` 无报错 |
| 1.2 | 拉取 ADK Go 依赖：`go get google.golang.org/adk@v1.0.0` | 15 | 1.1 | go.sum | 编译通过 |
| 1.3 | 拉取 LiteLLM 依赖：`go get github.com/BerriAI/litellm-go` | 15 | 1.2 | go.sum | `go vet` 无 warning |
| 1.4 | 创建 `cmd/opc-agent/` 空 main.go 骨架 | 10 | 1.1 | cmd/opc-agent/ | 目录存在 |
| 1.5 | 创建 `internal/{config,harness,memory,hitl,agent,runtime}/` 空包目录 | 15 | 1.1 | 目录结构 | `ls -R` 完整 |
| 1.6 | 创建 `plugins/{copywriter,email_sorter,xhs_poster}/` + `pkg/contracts/` + `build/plugins/` | 15 | 1.5 | 目录结构 | 所有目录存在 |
| 1.7 | 创建 `config.yaml` 模板（app/model/runtime/memory/harness 全字段） | 30 | 1.6 | config.yaml | YAML 格式合法 |

**WBS 1 合计：7 tasks / 1.75h**

---

## WBS 2 — `internal/config` 配置模块

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 2.1 | 定义 `Config` 顶层结构体 + yaml tag | 20 | 1.7 | config.go: Config | 编译通过 |
| 2.2 | 定义 `AppConfig` 子结构体 | 10 | 2.1 | config.go: AppConfig | 编译通过 |
| 2.3 | 定义 `ModelConfig` 子结构体 | 10 | 2.1 | config.go: ModelConfig | 编译通过 |
| 2.4 | 定义 `RuntimeConfig` 子结构体（PluginsDir） | 10 | 2.1 | config.go: RuntimeConfig | 编译通过 |
| 2.5 | 定义 `MemoryConfig` 子结构体（含 TopK/Threshold） | 10 | 2.1 | config.go: MemoryConfig | 编译通过 |
| 2.6 | 定义 `HarnessConfig` 子结构体 | 10 | 2.1 | config.go: HarnessConfig | 编译通过 |
| 2.7 | 实现 `Load(path)`：os.ReadFile → yaml.Unmarshal | 30 | 2.2-2.6 | config.go: Load | 合法 YAML 解析成功 |
| 2.8 | 实现 `Validate()`：必填检查 + 默认值填充 | 30 | 2.7 | config.go: Validate | 缺字段返回具体 error |
| 2.9 | 边缘场景处理：文件不存在使用默认值 + 格式错误返回 error | 20 | 2.8 | config.go: fallback | 各场景行为正确 |

**WBS 2 合计：9 tasks / 2.3h**

---

## WBS 3 — `internal/harness` Harness 控制

### 3a. RoleBoundary（角色边界）

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 3.1 | 定义 `RoleBoundary` 结构体 + `BoundaryViolationError` 错误类型 | 20 | — | role.go: struct | 编译通过 |
| 3.2 | 实现 `NewRoleBoundary()` 构造函数 | 10 | 3.1 | role.go: New | 返回实例 |
| 3.3 | 实现 `ValidateAction(action)`：在 AllowedTools 中 → 通过；不在 → 返回 error | 30 | 3.1 | role.go: ValidateAction | 合法/非法均正确 |
| 3.4 | 实现 `ValidateOutput(output)`：遍历 ForbiddenWords 检测 | 20 | 3.1 | role.go: ValidateOutput | 含禁止词→error |

### 3b. StateMachine（状态机）

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 3.5 | 定义 `AgentState` 枚举 + 所有状态常量 | 15 | — | state.go: enum | 编译通过 |
| 3.6 | 定义 `validTransitions` 转移矩阵（map[AgentState][]AgentState） | 20 | 3.5 | state.go: matrix | 覆盖所有状态 |
| 3.7 | 定义 `StateMachine` 结构体 + `NewStateMachine()` | 15 | 3.6 | state.go: struct | 编译通过 |
| 3.8 | 实现 `Transition(target)`：校验合法性 + 更新 current | 30 | 3.7 | state.go: Transition | 合法/非法转移正确 |
| 3.9 | 实现最大步数检测：transitions > maxTransitions → StateFailed | 15 | 3.8 | state.go: maxSteps | 超步数自动 Failed |
| 3.10 | 定义 `StateChangeListener` 接口 + `Subscribe()` | 15 | 3.7 | state.go: listener | 编译通过 |
| 3.11 | 实现 `notifyListeners()` + 默认控制台日志打印 | 20 | 3.10 | state.go: notify | 状态变更打印轨迹 |

### 3c. ArtifactContract（产物契约）

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 3.12 | 定义 `Artifact` 结构体（SchemaName/Version/Data/Raw/TokenUsage） | 15 | — | contract.go: struct | 编译通过 |
| 3.13 | 实现 `ValidateArtifact()`：SchemaName 匹配校验 | 20 | 3.12 | contract.go: Validate | 匹配/不匹配正确 |

### 3d. Guardrail（护栏规则）

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 3.14 | 定义 `GuardrailType`/`GuardrailAction` 枚举 + `GuardrailRule` 结构体 | 20 | — | guardrail.go: enum | 编译通过 |
| 3.15 | 定义 `GuardrailResult` + `GuardrailEngine` 结构体 | 15 | 3.14 | guardrail.go: struct | 编译通过 |
| 3.16 | 实现 `Check()`：遍历规则 → 命中判定 → 最高优先级 Action | 40 | 3.15 | guardrail.go: Check | Block > HITL > LogWarn |
| 3.17 | 实现 `ShouldBlock()` / `ShouldConfirm()` 便捷方法 | 20 | 3.16 | guardrail.go: helper | 正确返回 |

**WBS 3 合计：17 tasks / 5.8h**

---

## WBS 4 — `internal/memory` 记忆引擎

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 4.1 | 定义 `MemoryEntry` 结构体（ID/CreatedAt/TaskType/UserInput/AgentOutput/KeyDecisions/Embedding/Metadata） | 20 | — | entry.go: struct | 编译通过 |
| 4.2 | 定义 `MemoryStore` 接口（Store/Search/Recall/Count/Close） | 15 | 4.1 | store.go: interface | 编译通过 |
| 4.3 | 定义 `EmbeddedEngine` 结构体（entries/mu/storePath） | 15 | 4.2 | engine.go: struct | 编译通过 |
| 4.4 | 实现 `NewEmbeddedEngine(storePath)`：加载已有/创建空索引 | 20 | 4.3 | engine.go: New | 文件存在/不存在的处理 |
| 4.5 | 实现 `CosineSimilarity(a, b)`：纯 Go 数学计算 | 30 | — | engine.go: cosine | 相同→1.0, 正交→0.0 |
| 4.6 | 实现 `Store(ctx, entry)`：追加内存 + 原子写文件 | 30 | 4.4 | engine.go: Store | 写入后 JSON 合法 |
| 4.7 | 实现 `atomicPersist()`：临时文件→rename 覆盖 | 20 | 4.6 | engine.go: atomic | 崩溃安全 |
| 4.8 | 实现 `Search(ctx, query, topK)`：余弦相似度 + 阈值过滤 + TopK | 40 | 4.5, 4.6 | engine.go: Search | 返回正确排序的结果 |
| 4.9 | 实现 `Recall(ctx, text, topK)`：LLM embedding → Search | 30 | 4.8 | engine.go: Recall | Mock embedding 可测 |
| 4.10 | 实现 `Count(ctx)`：返回 entries 长度 | 5 | 4.4 | engine.go: Count | 正确计数 |
| 4.11 | 实现 `Close()`：空方法（预留） | 5 | 4.4 | engine.go: Close | 不 panic |
| 4.12 | 实现 `loadFromDisk()`：json.Unmarshal 恢复 | 20 | 4.4 | engine.go: load | 正常/空文件都可处理 |
| 4.13 | 实现文件损坏恢复：主文件坏→尝试 .bak 回退 | 20 | 4.12 | engine.go: backup | 损坏文件从 bak 恢复 |
| 4.14 | 实现 `NewEmptyEngine()` 降级策略 | 15 | 4.2 | engine.go: empty | 降级不 panic |
| 4.15 | 实现 `buildMemoryEntry()` 从任务结果构建记忆条目 | 20 | 4.1 | engine.go: builder | 字段填充正确 |

**WBS 4 合计：15 tasks / 5.3h**

---

## WBS 5 — `internal/hitl` HITL 模块

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 5.1 | 定义 `Operation` 结构体（Type/Description/Payload） | 10 | — | handler.go: struct | 编译通过 |
| 5.2 | 定义 `Handler` 结构体（reader/writer 接口注入） | 10 | — | handler.go: Handler | 编译通过 |
| 5.3 | 实现 `NewHandler(reader, writer)` 构造函数 | 10 | 5.2 | handler.go: New | 返回实例 |
| 5.4 | 实现 `Confirm(ctx, op)`：打印提示→bufio.Scanner 读入→判断 y/n | 40 | 5.3 | handler.go: Confirm | y→true, n→false, 其他→重新提示 |
| 5.5 | 实现可测试性：reader 支持 `bytes.Buffer` 注入 | 15 | 5.4 | handler.go: testable | Buffer 注入测试通过 |

**WBS 5 合计：5 tasks / 1.4h**

---

## WBS 6 — `internal/runtime` 插件运行时

### 6a. AgentPlugin 接口定义

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 6.1 | 定义 `PluginInfo` 结构体（Name/Summary/Version/Tags/RequiresHITL） | 15 | — | plugin.go: PluginInfo | 编译通过 |
| 6.2 | 定义 `ExecutionResult` 结构体（Data/TokenUsage） | 10 | — | plugin.go: ExecutionResult | 编译通过 |
| 6.3 | 定义 `TokenUsage` 结构体（InputTokens/OutputTokens） | 10 | — | plugin.go: TokenUsage | 编译通过 |
| 6.4 | 定义 `ReviewResult` 结构体（Passed/Score/CheckResults/Summary/ShouldRetry） | 10 | — | plugin.go: ReviewResult | 编译通过 |
| 6.5 | 定义 `AgentPlugin` 接口（Name/Info/Execute/Review） | 15 | 6.1-6.4 | plugin.go: AgentPlugin | 编译通过 |
| 6.6 | 定义 `PluginSymbol` 常量 = "Agent" | 5 | 6.5 | plugin.go: const | 编译通过 |

### 6b. Loader 实现

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 6.7 | 定义 `Loader` 结构体（pluginsDir/plugins map） | 10 | 6.5 | loader.go: struct | 编译通过 |
| 6.8 | 实现 `NewLoader(pluginsDir)` 构造函数 | 10 | 6.7 | loader.go: New | 返回实例 |
| 6.9 | 实现 `LoadAll()`：os.ReadDir → 过滤 .so → 逐个调用 loadFile | 30 | 6.8 | loader.go: LoadAll | 发现目录中所有 .so |
| 6.10 | 实现 `loadFile(soPath)`：plugin.Open → Lookup("Agent") → 类型断言 | 40 | 6.9 | loader.go: loadFile | 合法 .so 加载成功 |
| 6.11 | 实现名称唯一性检查 + 重复忽略 + 单个失败日志但不阻塞 | 20 | 6.10 | loader.go: dedup | 重复跳过，其余正常 |
| 6.12 | 实现 `Get(name)` 按名称查询 | 10 | 6.10 | loader.go: Get | 存在/不存在正确返回 |
| 6.13 | 实现 `List()` 返回所有 PluginInfo | 10 | 6.10 | loader.go: List | 返回完整列表 |
| 6.14 | 实现 `Count()` 返回数量 | 5 | 6.10 | loader.go: Count | 正确计数 |
| 6.15 | 插件约束文档化：Go 版本一致、依赖版本一致、平台限制 | 15 | 6.14 | — | 写入 README |

**WBS 6 合计：15 tasks / 3.6h**

---

## WBS 7 — `internal/agent` 内置 Agent

### 7a. Router Agent

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 7.1 | 定义 `RouteAction` 枚举 + `RouteResult` 结构体 | 15 | — | router.go: types | 编译通过 |
| 7.2 | 定义 `Router` 结构体（loader *runtime.Loader, model string） | 10 | 6.13 | router.go: struct | 编译通过 |
| 7.3 | 实现 `NewRouter(loader, model)` 构造函数 | 10 | 7.2 | router.go: New | 返回实例 |
| 7.4 | 实现 `buildL1Context(plugins)`：遍历 loader.List() 组装 prompt | 20 | 6.13 | router.go: L1 | 输出含所有插件名+简述 |
| 7.5 | 实现 `classifyPromptTemplate` 常量（含可用插件列表占位） | 15 | 7.4 | router.go: template | 模板正确 |
| 7.6 | 实现 `buildClassifyPrompt(input, plugins)`：模板填充 | 15 | 7.5 | router.go: buildPrompt | 输出含用户输入 |
| 7.7 | 实现 `classify(ctx, input)`：调 LLM → 解析 JSON | 30 | 7.6 | router.go: classify | Mock LLM 返回正确 agentName |
| 7.8 | 实现 `parseJSONResponse(raw)`：正则提取 JSON + json.Unmarshal | 25 | 7.7 | router.go: parseJSON | 合法/不合法 JSON 正确处理 |
| 7.9 | 实现 `Route(ctx, input)`：classify → 阈值判断 (0.4) → 匹配 → RouteResult | 30 | 7.7 | router.go: Route | 成功→Dispatch, 低置信→Unknown |
| 7.10 | 实现 unknown 兜底 + 用户提示构建 | 15 | 7.9 | router.go: unknown | 提示包含可用列表 |

### 7b. Reviewer Agent

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 7.11 | 定义 `ReviewResult` 结构体 + `CheckResult` 结构体 | 15 | — | reviewer.go: types | 编译通过 |
| 7.12 | 定义 `Reviewer` 结构体 + `NewReviewer(model)` | 10 | 7.11 | reviewer.go: struct | 编译通过 |
| 7.13 | 实现 `buildReviewPrompt(output, checkpoints)`：Agent 输出+检查清单→prompt | 20 | 7.12 | reviewer.go: buildPrompt | prompt 结构完整 |
| 7.14 | 实现 `Review(ctx, output, checkpoints)`：调 LLM → 解析 ReviewResult | 30 | 7.13 | reviewer.go: Review | Mock 审查返回正确结果 |
| 7.15 | 实现 `ReviewWithRetry(ctx, output, checkpoints, maxRetries)`：最多重试 1 次 | 25 | 7.14 | reviewer.go: retry | 重试后通过/不通过正确 |
| 7.16 | 实现审查不通过降级：打印审查报告 + 展示原始输出给用户 | 20 | 7.15 | reviewer.go: degrade | 控制台输出审查摘要 |
| 7.17 | 实现 `parseReviewResult(raw)`：LLM 审查响应结构化解析 | 20 | 7.14 | reviewer.go: parse | 结构化字段填充正确 |

### 7c. 公共工具

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 7.18 | 实现 `callLLM(ctx, model, prompt)` 公共 LLM 调用封装 | 30 | — | agent.go: callLLM | 返回响应 |
| 7.19 | 实现 `TokenTracker`：统计每次调用的 token 消耗 | 20 | 7.18 | agent.go: token | 累加计数正确 |
| 7.20 | 实现 `buildMemoryEntryFromResult(input, output, taskType)` | 15 | 4.1 | agent.go: memoryHelper | MemoryEntry 字段填充正确 |

**WBS 7 合计：20 tasks / 6.5h**

---

## WBS 8 — `plugins/copywriter` 文案插件

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 8.1 | 创建 `plugins/copywriter/go.mod`（独立 module） | 10 | 6.5 | go.mod | `go build` 成功 |
| 8.2 | 定义 `CopywriterPlugin` 空结构体 + `Name()` 返回 "copywriter" | 10 | 8.1 | plugin.go: struct | 编译通过 |
| 8.3 | 实现 `Info()`：返回 PluginInfo（标签/版本/HITL=true） | 15 | 8.2 | plugin.go: Info | 元数据正确 |
| 8.4 | 创建 `SKILL.md`：元数据 + L2 指令 + 输出格式 + 检查清单 + 禁止词 | 40 | 8.1 | SKILL.md | Markdown 结构完整 |
| 8.5 | 实现 `buildCopywritingPrompt(input, memories)`：组装 L2 指令+历史+用户输入 | 25 | 8.4 | plugin.go: buildPrompt | 输出含指令+记忆+用户输入 |
| 8.6 | 实现 `Execute(ctx, input, opts)`：调 LLM → 解析三段文案 | 40 | 8.5 | plugin.go: Execute | 输出符合 CopywritingOutput |
| 8.7 | 实现 `shortCopyValid(output)`：short ≤20 字 | 10 | 8.6 | plugin.go: validate | 超长报错 |
| 8.8 | 实现 `longCopyValid(output)`：long 100-200 字 | 10 | 8.6 | plugin.go: validate | 超范围报错 |
| 8.9 | 实现 `socialCopyValid(output)`：social ≤80 字 | 10 | 8.6 | plugin.go: validate | 超长报错 |
| 8.10 | 实现 `parseCopywritingOutput(raw)`：JSON 解析 + 校验 | 20 | 8.6 | plugin.go: parse | 合法/非法输出处理 |
| 8.11 | 导出 `var Agent CopywriterPlugin` 插件符号 | 5 | 8.2 | plugin.go: export | 编译为 .so |
| 8.12 | 编译 .so：`go build -buildmode=plugin -o ../../build/plugins/copywriter.so .` | 10 | 8.11 | copywriter.so | `file` 确认 ELF/Mach-O |

**WBS 8 合计：12 tasks / 3.4h**

---

## WBS 9 — `plugins/email_sorter` 邮件分类插件

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 9.1 | 创建 `plugins/email_sorter/go.mod` | 10 | 6.5 | go.mod | `go build` 成功 |
| 9.2 | 定义 `EmailSorterPlugin` + `Name()` + `Info()` | 15 | 9.1 | plugin.go: L1 | 元数据正确 |
| 9.3 | 创建 `SKILL.md`：分类标准（咨询/投诉/合作/垃圾）+ 回复建议格式 | 40 | 9.1 | SKILL.md | Markdown 结构完整 |
| 9.4 | 定义 `spamKeywords` 字符串数组（中文+英文关键词） | 15 | 9.2 | plugin.go: keywords | 覆盖常见垃圾词 |
| 9.5 | 实现 `quickSpamCheck(body)`：关键词预检快速判定 | 15 | 9.4 | plugin.go: spamCheck | 含关键词→true |
| 9.6 | 实现 `buildEmailPrompt(input, memories)`：L2 + 记忆 + 用户输入 | 20 | 9.3 | plugin.go: buildPrompt | prompt 结构正确 |
| 9.7 | 实现 `Execute(ctx, input, opts)`：调 LLM → 分类 → 生成回复建议 | 40 | 9.6 | plugin.go: Execute | 输出符合 EmailClassification |
| 9.8 | 实现紧急程度判定 `detectUrgency(body)`：基于关键词+语气规则 | 20 | 9.7 | plugin.go: urgency | 含紧急词→"高" |
| 9.9 | 实现 `parseEmailOutput(raw)`：JSON 解析 | 15 | 9.7 | plugin.go: parse | 字段映射正确 |
| 9.10 | 实现垃圾邮件标记逻辑：quickSpamCheck 通过→强制 category="垃圾" | 10 | 9.5 | plugin.go: spam | 命中关键词→垃圾 |
| 9.11 | 导出 `var Agent EmailSorterPlugin` | 5 | 9.2 | plugin.go: export | 编译为 .so |
| 9.12 | 编译 .so：go build -buildmode=plugin | 10 | 9.11 | email_sorter.so | 文件存在 |

**WBS 9 合计：12 tasks / 3.6h**

---

## WBS 10 — `plugins/xhs_poster` 小红书插件

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 10.1 | 创建 `plugins/xhs_poster/go.mod` | 10 | 6.5 | go.mod | `go build` 成功 |
| 10.2 | 定义 `XHSPosterPlugin` 结构体 + `Name()` 返回 "xhs_poster" | 10 | 10.1 | plugin.go: struct | 编译通过 |
| 10.3 | 实现 `Info()`：Tags=["social-media","xiaohongshu","content-marketing"]，RequiresHITL=true | 15 | 10.2 | plugin.go: Info | 元数据正确 |
| 10.4 | 创建 `SKILL.md`：小红书风格指令 + emoji 使用规范 + 图片描述要求 | 45 | 10.1 | SKILL.md | 包含完整 L2 指令 |
| 10.5 | 定义 XHS 特有词汇白名单（"种草"/"好物"/"亲测"/"安利" 等小红书常用语） | 15 | 10.2 | plugin.go: xhs_vocab | 覆盖主流词汇 |
| 10.6 | 实现 `buildXHSPrompt(input, memories)`：L2 + 小红书风格约束 + 用例 | 25 | 10.4 | plugin.go: buildPrompt | prompt 含风格约束 |
| 10.7 | 实现 `Execute(ctx, input, opts)`：调 LLM → 生成标题+正文+标签+配图描述 | 45 | 10.6 | plugin.go: Execute | 输出符合 XHSPost Schema |
| 10.8 | 实现标题校验 `validateTitle(title)`：含 emoji、≤20 字 | 10 | 10.7 | plugin.go: validateTitle | 非法标题报错 |
| 10.9 | 实现正文校验 `validateBody(body)`：200-500 字、含 emoji | 10 | 10.7 | plugin.go: validateBody | 超范围报错 |
| 10.10 | 实现标签推荐 `recommendHashtags(tags)`：5-10 个 + # 前缀确保 | 15 | 10.7 | plugin.go: hashtags | 标签格式正确 |
| 10.11 | 导出 `var Agent XHSPosterPlugin` | 5 | 10.2 | plugin.go: export | 编译为 .so |
| 10.12 | 编译 .so：go build -buildmode=plugin | 10 | 10.11 | xhs_poster.so | 文件存在 |

**WBS 10 合计：12 tasks / 3.6h**

---

## WBS 11 — `cmd/opc-agent` 主程序

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 11.1 | 实现 `parseFlags()`：--config + --plugins-dir 命令行参数 | 15 | — | main.go: flags | 默认值/自定义值正确 |
| 11.2 | 实现 Config Load + Validate + 错误处理 | 20 | 2.7 | main.go: init config | 配置无效→log.Fatal |
| 11.3 | 实现 Loader 初始化和 `LoadAll()` 调用 | 20 | 6.9 | main.go: init loader | 启动打印已加载插件数 |
| 11.4 | 实现 Memory Engine 初始化 + 降级处理 | 15 | 4.4, 4.14 | main.go: init memory | 启动时恢复记忆计数 |
| 11.5 | 实现 HITL Handler 初始化 | 10 | 5.3 | main.go: init hitl | 编译通过 |
| 11.6 | 实现 Router + Reviewer 初始化 | 15 | 7.3, 7.12 | main.go: init agents | 编译通过 |
| 11.7 | 实现 REPL 主循环：打印提示→scanner.Scan→processTask | 30 | 11.1-11.6 | main.go: REPL | 输入 exit 退出 |
| 11.8 | 实现 `processTask()` 全链路编排 | 45 | 11.7, 10.7, 7.9 | main.go: processTask | 端到端执行一次 |
| 11.9 | 实现超时控制：context.WithTimeout(30s) 包裹每个 task | 15 | 11.8 | main.go: timeout | 超时→打印提示 |
| 11.10 | 实现 Token 统计 + 任务结束后输出 | 15 | 11.8 | main.go: tokenStats | 统计信息打印 |
| 11.11 | 实现 Memory Store 写入回调（Agent 执行完后自动写入） | 20 | 11.8, 4.6 | main.go: memoryWrite | 任务结束时写入记忆 |

**WBS 11 合计：11 tasks / 3.7h**

---

## WBS 12 — 测试

### 12a. 单元测试

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 12.1 | Config 测试：正常加载 + 文件不存在 + 格式错误 + 缺字段 Validate | 40 | 2.x | config_test.go | 100% PASS |
| 12.2 | RoleBoundary 测试：合法工具+非法工具+禁止词命中+禁止词未命中 | 30 | 3.4 | role_test.go | 100% PASS |
| 12.3 | StateMachine 测试：合法转移+非法转移+超步数+监听器通知 | 40 | 3.11 | state_test.go | 覆盖全部状态 |
| 12.4 | Guardrail 测试：单规则命中+多规则优先级+无规则命中 | 30 | 3.16 | guardrail_test.go | Block > HITL > Warn |
| 12.5 | Memory 测试：Store+Search+Recall+Count+原子写+文件恢复+降级 | 60 | 4.x | engine_test.go | 覆盖正常/异常 |
| 12.6 | HITL 测试：y/n/空输入/reader 接口注入 | 20 | 5.x | handler_test.go | 100% PASS |
| 12.7 | Loader 测试：空目录/有 .so/无 .so/重复名/单个失败 | 40 | 6.x | loader_test.go | 各场景正确 |
| 12.8 | Router 测试：正常分发/unknown/低置信度/LLM 超时/JSON 解析失败 | 50 | 7.10 | router_test.go | Mock LLM 全覆盖 |
| 12.9 | Reviewer 测试：审查通过/不通过/重试后通过/重试后仍不通过 | 40 | 7.16 | reviewer_test.go | 重试逻辑正确 |

### 12b. 集成测试

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 12.10 | 创建测试用 dummy .so 插件（最小化 AgentPlugin 实现） | 30 | 6.x | testdata/dummy/ | Loader 可加载 |
| 12.11 | 编写端到端集成测试：Router→Plugin→Reviewer→HITL→Memory 全链路 Mock | 60 | 12.10, 11.8 | integration_test.go | 全链路通过 |
| 12.12 | 编写多插件加载测试：同时加载 3 个 .so + Router 发现所有 | 30 | 12.10, 8.x, 9.x, 10.x | loader_integration_test.go | 3 个全部加载 |

### 12c. 非功能测试

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 12.13 | NF1 响应时间测试：20 次 Mock 任务，单次 < 3s（不含 LLM） | 30 | 11.8 | nf_test.go | 全部 < 3s |
| 12.14 | NF2 成功率测试：连续 20 次任务，成功率 ≥ 90% | 30 | 11.8 | nf_test.go | ≤ 2 次失败 |
| 12.15 | NF3 路由准确率测试：30 次不同输入，准确率 ≥ 90% | 30 | 7.9 | nf_test.go | ≤ 3 次错误 |
| 12.16 | NF4 Token 节省率测试：对比全量注入 vs Skill 渐进式 | 40 | 8.x | nf_test.go | 节省 ≥ 60% |
| 12.17 | NF5 扩展性测试：新增一个插件只需编译该插件，不动主程序 | 20 | 10.x | nf_test.go | 主程序不重新编译 |
| 12.18 | NF7 内存占用测试：执行 100 次任务后 RSS < 500MB | 30 | 11.8 | nf_test.go | RSS < 500MB |

**WBS 12 合计：18 tasks / 10.5h**

---

## WBS 13 — 工程化与文档

| WBS | 任务 | 工时(m) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| 13.1 | 编写 README.md：项目简介 + 架构 + 快速开始 + 插件开发指南 | 40 | 11.x | README.md | 包含构建/运行步骤 |
| 13.2 | 编写 `pkg/contracts/*.json` 三个 JSON Schema 文件 | 30 | 8/9/10.x | copies JSON | Schema 语法合法 |
| 13.3 | 完善 Makefile：build-main/build-plugins/test/test-all/lint/clean/run | 40 | 11.x | Makefile | `make test-all` 通过 |
| 13.4 | 编译所有插件 + 主程序 + `go vet ./...` 零 error | 30 | 13.3 | 二进制产物 | `go vet` 全部通过 |
| 13.5 | 编写 MVP 验证报告模板（测试数据 + 假设验证 + 优化建议） | 30 | 12.x | MVP_VERIFICATION.md | 模板结构完整 |
| 13.6 | 安装 `golangci-lint` + 第一轮 lint 修复 | 20 | 13.4 | — | lint 零 warning |

**WBS 13 合计：6 tasks / 3.2h**

---

## 工时汇总

| WBS | 工作包 | 任务数 | 工时(h) | 占比 |
|-----|--------|--------|---------|------|
| 1 | 项目启动与环境搭建 | 7 | 1.75 | 2.1% |
| 2 | internal/config 配置模块 | 9 | 2.3 | 2.8% |
| 3 | internal/harness Harness 控制 | 17 | 5.8 | 7.0% |
| 4 | internal/memory 记忆引擎 | 15 | 5.3 | 6.4% |
| 5 | internal/hitl HITL 模块 | 5 | 1.4 | 1.7% |
| 6 | internal/runtime 插件运行时 | 15 | 3.6 | 4.3% |
| 7 | internal/agent 内置 Agent | 20 | 6.5 | 7.8% |
| 8 | plugins/copywriter 文案插件 | 12 | 3.4 | 4.1% |
| 9 | plugins/email_sorter 邮件插件 | 12 | 3.6 | 4.3% |
| 10 | plugins/xhs_poster 小红书插件 | 12 | 3.6 | 4.3% |
| 11 | cmd/opc-agent 主程序 | 11 | 3.7 | 4.4% |
| 12 | 测试 | 18 | 10.5 | 12.6% |
| 13 | 工程化与文档 | 6 | 3.2 | 3.8% |
| **总计** | | **159** | **~50h** | 100% |

---

## 并行执行图

```
Day  1   2   3   4   5   6   7   8   9  10  11  12  13  14
    ┌───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───┐
WBS1│███│   │   │   │   │   │   │   │   │   │   │   │   │   │
WBS2│   │███│██ │   │   │   │   │   │   │   │   │   │   │   │
WBS3│   │   │███│███│██ │   │   │   │   │   │   │   │   │   │
WBS4│   │   │   │███│███│██ │   │   │   │   │   │   │   │   │
WBS5│   │   │   │██ │   │   │   │   │   │   │   │   │   │   │
WBS6│   │   │   │   │███│███│   │   │   │   │   │   │   │   │
WBS7│   │   │   │   │   │███│███│██ │   │   │   │   │   │   │
WBS8│   │   │   │   │   │   │   │███│███│   │   │   │   │   │
WBS9│   │   │   │   │   │   │   │   │███│███│   │   │   │   │
WBS10│   │   │   │   │   │   │   │   │   │███│███│   │   │   │
WBS11│   │   │   │   │   │   │   │   │   │   │███│███│   │   │
WBS12│   │   │   │   │   │   │   │   │   │   │   │███│███│██ │
WBS13│   │   │   │   │   │   │   │   │   │   │   │   │   │███│
    └───┴───┴───┴───┴───┴───┴───┴───┴───┴───┴───┴───┴───┴───┘
```

> **并行策略**：Core 模块 (WBS 2-5) 和插件模块 (WBS 8-10) 内部可并行开发；wbs 6 (Runtime) 是所有插件的前置依赖。

---

## 里程碑与交付物

| 里程碑 | 时间 | 交付物 | 关联 WBS |
|--------|------|--------|----------|
| **M0：环境就绪** | 第 1 天 | go.mod + 目录骨架 + Makefile + config.yaml | WBS 1 |
| **M1：Core 框架完成** | 第 4 天 | Config + Harness + Memory + HITL + 单元测试 | WBS 2-5 |
| **M2：插件运行时完成** | 第 6 天 | AgentPlugin 接口 + Loader + 测试用 dummy 插件 | WBS 6 |
| **M3：内置 Agent 完成** | 第 8 天 | Router + Reviewer + 单元测试 | WBS 7 |
| **M4：业务插件完成** | 第 12 天 | copywriter.so + email_sorter.so + xhs_poster.so | WBS 8-10 |
| **M5：端到端闭环** | 第 12 天 | REPL + 全链路编排 + 冒烟通过 | WBS 11 |
| **M6：测试完成** | 第 14 天 | 全部 UT/集成/NF 测试 PASS | WBS 12 |
| **M7：MVP 发布** | 第 14 天 | README + 二进制产物 + MVP 验证报告 | WBS 13 |

---

*文档结束 — v2.0 含 134 项细化任务 + 小红书 Agent 插件*
