# OPC-Agent WBS 工作分解结构

> **版本**：v1.0  
> **日期**：2026-05-10  
> **关联文档**：[需求说明书 v1.1](./01%20多任务智能协助系统（OPC-Agent）——项目需求说明书.md)，[详细设计说明书](./07%20OPC-Agent%20详细设计说明书.md)  
> **项目代号**：OPC-Agent  
> **总工期**：14 天（MVP）  
> **总任务数**：52 项

---

## WBS 概览

```
OPC-Agent MVP
│
├── 1. 项目启动与环境搭建 (WBS 1)             第 1 天
│   ├── 1.1  Go 模块初始化
│   ├── 1.2  目录结构搭建
│   ├── 1.3  配置文件模板
│   └── 1.4  构建脚本与 Makefile
│
├── 2. Core 框架层 (WBS 2)                   第 2-4 天
│   ├── 2.1  internal/config — 配置模块
│   ├── 2.2  internal/harness — Harness 控制
│   ├── 2.3  internal/memory — 记忆引擎
│   └── 2.4  internal/hitl — HITL 模块
│
├── 3. Plugin 运行时 (WBS 3)                  第 4-5 天
│   ├── 3.1  AgentPlugin 接口定义
│   ├── 3.2  Loader 插件加载器
│   └── 3.3  插件 SDK 文档
│
├── 4. 内置 Agent (WBS 4)                    第 5-6 天
│   ├── 4.1  Router Agent
│   └── 4.2  Reviewer Agent
│
├── 5. Agent 插件开发 (WBS 5)                第 6-9 天
│   ├── 5.1  copywriter 插件
│   └── 5.2  email_sorter 插件
│
├── 6. 主程序与流程编排 (WBS 6)               第 9-10 天
│   ├── 6.1  main.go 入口
│   ├── 6.2  REPL 交互循环
│   └── 6.3  全链路编排
│
├── 7. 测试 (WBS 7)                          第 11-12 天
│   ├── 7.1  单元测试
│   ├── 7.2  集成测试
│   └── 7.3  非功能测试
│
├── 8. 工程化与文档 (WBS 8)                  第 13-14 天
│   ├── 8.1  README + 开发指南
│   ├── 8.2  Makefile 完善
│   └── 8.3  MVP 验证报告
│
└── 9. 里程碑与交付物
```

---

## WBS 1 — 项目启动与环境搭建

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 1.1 | 初始化 Go module：`go mod init github.com/yourname/opc-agent` | 0.5 | — | go.mod, go.sum | `go build ./...` 无报错 |
| 1.2 | 创建目录骨架：`cmd/opc-agent/`、`internal/{config,harness,memory,hitl,agent,runtime}/`、`plugins/{copywriter,email_sorter}/`、`build/`、`build/plugins/`、`pkg/contracts/` | 0.5 | 1.1 | 空目录结构 | `ls -R` 确认结构完整 |
| 1.3 | 创建 `config.yaml` 模板：包含 app/model/runtime/memory/harness 配置项 | 1 | 1.1 | config.yaml | `config.Load()` 解析正常 |
| 1.4 | 编写 Makefile：target 包括 `build-main`、`build-plugin`、`test`、`clean`、`run` | 1 | 1.2 | Makefile | `make build-main` 编译成功 |

**工时小计：3h**

---

## WBS 2 — Core 框架层

### WBS 2.1 `internal/config` — 配置模块

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 2.1.1 | 定义 `Config` 结构体（App + Model + Runtime + Memory + Harness） | 1 | 1.3 | config.go: Config 定义 | 编译通过 |
| 2.1.2 | 实现 `Load(path string) (*Config, error)`：YAML 读取 + `yaml.Unmarshal` | 1 | 2.1.1 | Load 函数 | 传入合法/非法 YAML，验证结果 |
| 2.1.3 | 实现 `Validate() error`：必填字段检查 + 默认值填充 | 1 | 2.1.2 | Validate 方法 | 缺少必填字段时返回具体 error |
| 2.1.4 | 处理 Config 文件不存在、格式错误等边缘场景 | 0.5 | 2.1.3 | 错误处理 | 文件不存在时用默认值（warn） |
| 2.1.5 | 为 `RuntimeConfig` 和 `MemoryConfig` 编写单元测试 | 1 | 2.1.4 | config_test.go | `go test` 全部 PASS |

**工时小计：4.5h**

### WBS 2.2 `internal/harness` — Harness 控制系统

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 2.2.1 | 实现 `RoleBoundary`：结构体定义 + `ValidateAction()` + `ValidateOutput()` | 1.5 | — | role.go | 合法/非法操作校验通过 |
| 2.2.2 | 实现 `StateMachine`：状态定义 + 转移矩阵 + `Transition()` + 非法转移检测 | 2 | — | state.go | 合法/非法转移均正确响应 |
| 2.2.3 | 实现状态变更监听器 + 控制台打印日志 | 1 | 2.2.2 | state.go: Subscribe, notify | 状态变更时打印正确轨迹 |
| 2.2.4 | 实现最大步数检测（死循环防护） | 0.5 | 2.2.2 | state.go: maxTransitions | 超过步数时自动变为 Failed |
| 2.2.5 | 实现 `Artifact` 结构体 + `ValidateArtifact()` 契约校验 | 1 | — | contract.go | Schema 匹配/不匹配均正确 |
| 2.2.6 | 实现 `GuardrailEngine`：规则定义 + `Check()` 多规则优先级 | 2 | — | guardrail.go | Block > HITLConfirm > LogWarn |
| 2.2.7 | 为所有 harness 模块编写单元测试 | 2 | 2.2.1~2.2.6 | *_test.go | `go test -cover` 覆盖率 > 80% |

**工时小计：10h**

### WBS 2.3 `internal/memory` — 记忆引擎

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 2.3.1 | 定义 `MemoryEntry` 结构体 + `MemoryStore` 接口（Store/Search/Recall/Count/Close） | 1 | — | store.go, entry.go | 编译通过 |
| 2.3.2 | 实现 `CosineSimilarity()` 向量余弦相似度计算 | 1 | 2.3.1 | engine.go: CosineSimilarity | 相同向量 → 1.0，正交向量 → 0.0 |
| 2.3.3 | 实现 `EmbeddedEngine.Store()`：追加内存索引 + 原子写文件 | 2 | 2.3.1 | engine.go: Store | 写入后文件存在，内容合法 JSON |
| 2.3.4 | 实现 `EmbeddedEngine.Search()`：余弦相似度 + TopK + 阈值过滤 | 2 | 2.3.2, 2.3.3 | engine.go: Search | 返回 TopK 且相似度 > 阈值 |
| 2.3.5 | 实现 `EmbeddedEngine.Recall()`：输入文本 → 调用 embedding API → Search | 1.5 | 2.3.4 | engine.go: Recall | Mock embedding 返回正确结果 |
| 2.3.6 | 实现 `loadFromDisk()` 启动恢复 + 文件损坏时的 .bak 回退 | 1.5 | 2.3.3 | engine.go: loadFromDisk | 正常/损坏文件均正确处理 |
| 2.3.7 | 实现 `NewEmptyEngine()` 降级策略：Search 返回空，Store 静默丢弃 | 0.5 | 2.3.1 | engine.go 降级 | 降级模式下不 panic |
| 2.3.8 | 为 memory 模块编写完整单元测试 | 2.5 | 2.3.1~2.3.7 | engine_test.go | 覆盖正常/异常/边缘路径 |

**工时小计：12h**

### WBS 2.4 `internal/hitl` — HITL 模块

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 2.4.1 | 定义 `Operation` 结构体 + `Handler` 结构体 | 0.5 | — | handler.go: 类型定义 | 编译通过 |
| 2.4.2 | 实现 `Confirm()`：终端打印提示 → 读入 y/n → 返回布尔值 | 1.5 | 2.4.1 | handler.go: Confirm | 输入 y → true，n → false |
| 2.4.3 | 实现可测试性设计：`reader`/`writer` 接口注入（支持 bytes.Buffer） | 0.5 | 2.4.2 | handler.go: 注入 | 使用 bytes.Buffer 测试通过 |
| 2.4.4 | 编写 HITL 单元测试 | 0.5 | 2.4.3 | handler_test.go | `go test` 覆盖 y/n/空输入 |

**工时小计：3h**

---

## WBS 3 — Plugin 运行时

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 3.1 | 定义 `AgentPlugin` 接口（Name/Info/Execute/Review）+ `PluginInfo` + `ExecutionResult` | 1.5 | — | plugin.go | 编译通过 |
| 3.2 | 定义 `TokenUsage` + `ReviewResult` 等配套类型 | 0.5 | 3.1 | plugin.go: 类型 | 编译通过 |
| 3.3 | 实现 `Loader` 结构体 + `NewLoader(pluginsDir)` | 0.5 | 3.1 | loader.go: 结构体 | 编译通过 |
| 3.4 | 实现 `LoadAll()`：扫描目录 → 过滤 .so → 逐个调用 `loadFile()` | 2 | 3.3 | loader.go: LoadAll | 正确发现 plugins 目录中所有 .so |
| 3.5 | 实现 `loadFile()`：`plugin.Open()` → `Lookup("Agent")` → 类型断言 → 注册 | 2 | 3.4 | loader.go: loadFile | 合法 .so 成功加载，非法文件报错 |
| 3.6 | 实现名称唯一性检查 + 重复忽略 + 单个失败不阻塞其余 | 1 | 3.5 | loader.go: 去重 | 重复名称跳过，其他插件正常加载 |
| 3.7 | 实现 `Get(name)` / `List()` / `Count()` 查询方法 | 0.5 | 3.5 | loader.go: 查询 | 正确返回已加载插件信息 |
| 3.8 | 为 runtime 模块编写单元测试（含 mock .so / 接口 Mock） | 2 | 3.1~3.7 | loader_test.go | `go test -race` 无竞争 |

**工时小计：10h**

---

## WBS 4 — 内置 Agent

### WBS 4.1 Router Agent

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 4.1.1 | 定义 `Router` 结构体 + `NewRouter(loader, model)` | 0.5 | 3.1 | router.go: 初始化 | 编译通过 |
| 4.1.2 | 实现 `buildClassifyPrompt()`：从 Loader.List() 构建 L1 上下文 | 1 | 4.1.1 | router.go: prompt 构建 | 输出正确包含所有插件名+简述 |
| 4.1.3 | 实现 `classify()`：调用 LLM → 解析 JSON → 返回 (agentName, confidence) | 2 | 4.1.2 | router.go: 意图识别 | Mock LLM 返回对应结果 |
| 4.1.4 | 实现 `Route()`：classify → 阈值判断 → 匹配插件 → RouteResult | 1.5 | 4.1.3 | router.go: 分发逻辑 | 识别成功返回插件，unknown 返回提示 |
| 4.1.5 | 实现信心度 < 0.4 时的 unknown 兜底 + 重试机制 | 1 | 4.1.4 | router.go: 兜底 | 低置信度正确返回 unknown |
| 4.1.6 | 编写 Router 单元测试 | 2 | 4.1.1~4.1.5 | router_test.go | Mock LLM 覆盖 成功/unknown/超时 |

**工时小计：8h**

### WBS 4.2 Reviewer Agent

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 4.2.1 | 定义 `Reviewer` 结构体 + `NewReviewer(model)` | 0.5 | — | reviewer.go | 编译通过 |
| 4.2.2 | 实现 `buildReviewPrompt()`：Agent 输出 + 检查清单 → LLM 审查 prompt | 1.5 | 4.2.1 | reviewer.go: prompt | 输出包含检查项和输出内容 |
| 4.2.3 | 实现 `Review()`：调用 LLM → 解析 `ReviewResult` → 返回 | 2 | 4.2.2 | reviewer.go: 审查 | Mock LLM 审查结果正确解析 |
| 4.2.4 | 实现 `ReviewWithRetry()`：最多 1 次重试 → 修正后重审 | 1.5 | 4.2.3 | reviewer.go: 重试 | 重试后通过 / 重试后仍不通过 |
| 4.2.5 | 实现审查不通过时的降级策略：展示用户，由用户决定 | 1 | 4.2.4 | reviewer.go: 降级 | 输出未通过原因和原始结果 |
| 4.2.6 | 编写 Reviewer 单元测试 | 1.5 | 4.2.1~4.2.5 | reviewer_test.go | `go test` 全部 PASS |

**工时小计：8h**

---

## WBS 5 — Agent 插件开发

### WBS 5.1 copywriter 插件

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 5.1.1 | 创建 `plugins/copywriter/go.mod`（依赖主程序的 runtime 包） | 0.5 | 3.1 | go.mod | `go build` 成功 |
| 5.1.2 | 编写 `CopywriterPlugin` 结构体 + `Name()` / `Info()` 实现 | 0.5 | 5.1.1 | plugin.go: L1 | 返回正确元数据 |
| 5.1.3 | 在插件内创建 `SKILL.md`：包含元数据、L2 指令、检查清单、禁止词 | 1.5 | 5.1.1 | SKILL.md | Markdown 格式正确 |
| 5.1.4 | 实现 `Execute()`：调用 LLM + 解析三段式文案（短/长/社交） | 3 | 5.1.2 | plugin.go: Execute | 输出符合 CopywritingOutput Schema |
| 5.1.5 | 实现输出契约验证：short ≤20、long 100-200、social ≤80 | 1 | 5.1.4 | plugin.go: validate | 超长/超短均正确报错 |
| 5.1.6 | 实现 HITL 集成：在 Execute 结果中标记 RequiresHITL=true | 0.5 | 5.1.4 | plugin.go: HITL | Info().RequiresHITL == true |
| 5.1.7 | 编译为 .so：`go build -buildmode=plugin -o ../../build/plugins/copywriter.so .` | 0.5 | 5.1.1~5.1.6 | copywriter.so | `file copywriter.so` 显示 ELF/Mach-O |
| 5.1.8 | 为主程序的 Loader 加载测试：将 copywriter.so 放入 plugins 目录 | 0.5 | 5.1.7 | 集成验证 | Loader.LoadAll() 成功加载 |

**工时小计：8h**

### WBS 5.2 email_sorter 插件

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 5.2.1 | 创建 `plugins/email_sorter/go.mod` | 0.5 | 3.1 | go.mod | `go build` 成功 |
| 5.2.2 | 编写 `EmailSorterPlugin` 结构体 + `Name()` / `Info()` 实现 | 0.5 | 5.2.1 | plugin.go: L1 | 返回正确元数据 |
| 5.2.3 | 创建 `SKILL.md`：分类标准（咨询/投诉/合作/垃圾）、回复建议格式 | 1.5 | 5.2.1 | SKILL.md | Markdown 格式正确 |
| 5.2.4 | 实现 `quickSpamCheck()` 关键词预检 | 1 | 5.2.2 | plugin.go: 垃圾检测 | 含关键词 → true，不含 → false |
| 5.2.5 | 实现 `Execute()`：LLM 分类 + 生成 reason + reply_suggestion | 3 | 5.2.2 | plugin.go: Execute | 输出符合 EmailClassification Schema |
| 5.2.6 | 实现紧急程度标记：基于关键词+语气识别高/中/低 | 1.5 | 5.2.5 | plugin.go: urgency | 含紧急词 → "高" |
| 5.2.7 | 编译为 .so：`go build -buildmode=plugin -o ../../build/plugins/email_sorter.so .` | 0.5 | 5.2.1~5.2.6 | email_sorter.so | .so 文件存在 |
| 5.2.8 | Loader 集成验证 | 0.5 | 5.2.7 | 集成验证 | 两个插件同时加载成功 |

**工时小计：9h**

---

## WBS 6 — 主程序与流程编排

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 6.1 | 编写 `main.go`：flag 解析 → Load Config → Loader.LoadAll → 初始化 Memory → HITL → Router → Reviewer | 2 | 2.1, 3.x, 4.x | cmd/opc-agent/main.go | 启动时打印插件加载信息 |
| 6.2 | 实现 REPL 主循环：读取 stdin → `processTask()` → 输出结果 | 1.5 | 6.1 | main.go: REPL | 输入 exit 退出，输入指令执行 |
| 6.3 | 实现 `processTask()`：Router.Route → Plugin.Execute → Reviewer.Review → HITL → Memory.Store | 3 | 6.1, 4.x, 5.x | main.go: 编排 | 全链路执行一次 |
| 6.4 | 实现超时控制：每个任务绑定 `context.WithTimeout(30s)` | 1 | 6.3 | main.go: 超时 | 超过 30s 打印超时提示 |
| 6.5 | 实现 Token 统计与输出：任务结束时打印 input/output tokens | 1 | 6.3 | main.go: 统计 | 控制台显示 Token 消耗 |
| 6.6 | 实现重试与降级矩阵（见详细设计第 4.3 节） | 1.5 | 6.3 | main.go: 降级 | 各失败场景正确降级 |
| 6.7 | 实现 `memoryRetrievalTool()` 记忆检索注入插件 opts | 1 | 6.3, 2.3 | main.go: 记忆注入 | 插件 Execute 时 opts 含 memories |
| 6.8 | 编译主程序：`go build -o ../../build/opc-agent .` | 0.5 | 6.1~6.7 | build/opc-agent | 二进制文件存在 |
| 6.9 | 全链路冒烟测试：启动 → 输入"写文案" → 得到输出 | 1 | 6.8 | 冒烟通过 | 端到端拿到输出 |

**工时小计：12.5h**

---

## WBS 7 — 测试

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 7.1 | 编写 Config 单元测试（加载/校验/默认值/错误） | 1 | 2.1 | config_test.go | 100% PASS |
| 7.2 | 编写 Harness 单元测试（Role/State/Contract/Guardrail） | 2 | 2.2 | harness/*_test.go | 覆盖状态转移矩阵 |
| 7.3 | 编写 Memory 单元测试（Store/Search/Recall/持久化/恢复） | 2 | 2.3 | memory/*_test.go | 覆盖文件损坏恢复 |
| 7.4 | 编写 HITL 单元测试（y/n/空输入/超时） | 0.5 | 2.4 | hitl/handler_test.go | 100% PASS |
| 7.5 | 编写 Runtime 单元测试（Loader 发现/加载/去重/错误） | 1.5 | 3.x | runtime/*_test.go | 单个失败不阻塞 |
| 7.6 | 编写 Router 单元测试（分类/分发/unknown/超时） | 1.5 | 4.1 | agent/router_test.go | Mock LLM 全场景 |
| 7.7 | 编写 Reviewer 单元测试（审查/重试/降级） | 1 | 4.2 | agent/reviewer_test.go | 重试逻辑正确 |
| 7.8 | 编写插件 Mock：创建测试用 dummy .so 插件 | 1.5 | 3.x | testdata/dummy_plugin/ | Loader 可加载 |
| 7.9 | 编写集成测试：用 dummy 插件执行端到端流程 | 2 | 7.8, 6.3 | integration_test.go | 全链路 Mock 通过 |
| 7.10 | NF 测试：20 次连续任务成功率 ≥ 90% | 1 | 6.8, 7.9 | nf_test.go | ≤ 2 次失败 |
| 7.11 | NF 测试：路由准确率 30 次 ≥ 90% | 1 | 7.10 | nf_test.go | ≤ 3 次错误 |
| 7.12 | NF 测试：内存占用 ≤ 500MB（执行 100 次后检查 RSS） | 1 | 6.8 | nf_test.go | RSS < 500MB |

**工时小计：16h**

---

## WBS 8 — 工程化与文档

| WBS | 任务 | 工时(h) | 前置依赖 | 产出物 | QA 验证方式 |
|-----|------|---------|----------|--------|-------------|
| 8.1 | 编写项目 README.md：简介、架构、快速开始、构建命令 | 1 | 6.8 | README.md | README 包含构建和运行步骤 |
| 8.2 | 完善 Makefile：增加 `lint`、`clean-all`、`run`、`test-all` 目标 | 1 | 6.8 | Makefile | `make test-all` 全部通过 |
| 8.3 | 编写 MVP 验证报告模板 | 1 | — | MVP_VERIFICATION.md | 包含测试数据和结论模板 |
| 8.4 | 编译所有插件 + 主程序，确认零 error | 1 | 5.x, 6.8 | 构建产物 | `go vet ./...` 无 warning |

**工时小计：4h**

---

## WBS 9 — 里程碑与交付物

| 里程碑 | 时间 | 交付物 | 关联 WBS |
|--------|------|--------|----------|
| **M0：环境就绪** | 第 1 天 | go.mod + 空目录骨架 + Makefile + config.yaml | WBS 1 |
| **M1：Core 框架完成** | 第 4 天 | Config + Harness + Memory + HITL 全部实现并通过单元测试 | WBS 2 |
| **M2：插件运行时完成** | 第 5 天 | AgentPlugin 接口 + Loader 实现 + 单元测试 | WBS 3 |
| **M3：内置 Agent 完成** | 第 6 天 | Router + Reviewer 实现 + 单元测试 | WBS 4 |
| **M4：业务插件完成** | 第 9 天 | copywriter.so + email_sorter.so 可被 Loader 加载 | WBS 5 |
| **M5：端到端闭环** | 第 10 天 | main.go REPL + 全链路编排 + 冒烟测试通过 | WBS 6 |
| **M6：测试完成** | 第 12 天 | 全部单元 + 集成 + NF 测试 PASS | WBS 7 |
| **M7：MVP 发布** | 第 14 天 | README + 构建产物 + MVP 验证报告 | WBS 8 |

---

## 工时汇总

| WBS | 工作包 | 工时(h) | 占比 |
|-----|--------|---------|------|
| 1 | 项目启动与环境搭建 | 3 | 3.3% |
| 2 | Core 框架层 | 29.5 | 32.8% |
| 3 | Plugin 运行时 | 10 | 11.1% |
| 4 | 内置 Agent | 16 | 17.8% |
| 5 | Agent 插件开发 | 17 | 18.9% |
| 6 | 主程序与流程编排 | 12.5 | 13.9% |
| 7 | 测试 | 16 | 17.8% |
| 8 | 工程化与文档 | 4 | 4.4% |
| **总计** | | **90h** | **100%** |

> **注**：测试工时（WBS 7, 16h）已分摊到各模块工时中。上表 WBS 2~6 不含测试，WBS 7 为额外的端到端和 NF 测试。

---

*文档结束*
