# OPC-Agent 内容营销升级 WBS 工作分解结构

> **版本**：v1.0
> **日期**：2026-05-13
> **关联文档**：
> - [12 内容营销升级需求说明书](./12%20OPC-Agent%20%E5%86%85%E5%AE%B9%E8%90%A5%E9%94%80%E5%8D%87%E7%BA%A7%E9%9C%80%E6%B1%82%E8%AF%B4%E6%98%8E%E4%B9%A6.md)
> - [13 内容创作增强层需求说明书（Phase A）](./13%20OPC-Agent%20%E5%86%85%E5%AE%B9%E5%88%9B%E4%BD%9C%E5%A2%9E%E5%BC%BA%E5%B1%82%E9%9C%80%E6%B1%82%E8%AF%B4%E6%98%8E%E4%B9%A6%EF%BC%88Phase%20A%EF%BC%89.md)
> - [14 商业化变现层需求说明书（Phase C）](./14%20OPC-Agent%20%E5%95%86%E4%B8%9A%E5%8C%96%E5%8F%98%E7%8E%B0%E5%B1%82%E9%9C%80%E6%B1%82%E8%AF%B4%E6%98%8E%E4%B9%A6%EF%BC%88Phase%20C%EF%BC%89.md)
> - [15 内容营销升级用户故事文档](./15%20OPC-Agent%20%E5%86%85%E5%AE%B9%E8%90%A5%E9%94%80%E5%8D%87%E7%BA%A7%E7%94%A8%E6%88%B7%E6%95%85%E4%BA%8B%E6%96%87%E6%A1%A3.md)
> - [08 WBS 工作分解结构（MVP）](./08%20OPC-Agent%20WBS%E5%B7%A5%E4%BD%9C%E5%88%86%E8%A7%A3%E7%BB%93%E6%9E%84.md)
> **前置 MVP**：v0.2.0（已完成：CLI + Web UI + 桌面端 + SaaS + Workflow）
> **项目代号**：OPC-Agent Content Marketing
> **总工期**：Phase B 8周 + Phase A 3周 + Phase C 16周 = **27周**
> **总任务数**：**240 项**

---

## 执行策略

### 波次依赖图

```
Phase B ─────────────────────────────────────────────────────
Wave B1 (Week 1-2): 平台 SDK 框架 + 第一个平台适配 + MCP + DB
   │
   ├──→ Wave B2 (Week 2-4): Publish 引擎 + 发布 API + UI
   │         │
   │         └──→ Wave B3 (Week 4-6): Engage 引擎 + 互动 UI
   │
   └──→ Wave B4 (Week 6-8): Create 增强 + MCP 完善 + 集成

Phase A (Week 9-11) ────────────────────────────────────────
Wave A1 (Week 9):   video_script Agent
Wave A2 (Week 10):  tag_generator Agent
Wave A3 (Week 11):  trend_radar Agent

Phase C (Week 12-27) ───────────────────────────────────────
Wave C1 (Week 12-15): Marketplace 核心 + DB + API
Wave C2 (Week 16-19): 订单流 + 交付 + 审核
Wave C3 (Week 20-23): 结算引擎 + 钱包 + 提现
Wave C4 (Week 24-27): 评价体系 + 数据报告 + 反作弊 + 集成
```

### 并行原则

**同一 Wave 内的任务可以并行执行**（除非标注了前置依赖）。多个人可以同时在 Wave 内的不同任务上工作。

---

## Phase B — 内容平台核心层（WBS 1-4）

总计：**4 个 Wave，113 项任务，8 周**

---

### WBS 1 — Wave B1：平台 SDK 基础 + 账号体系 + MCP 底座（Week 1-2）

**目标**：搭建平台 SDK 框架、完成首个平台适配（抖音）、MCP Server 骨架、数据库扩展

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **1.1 平台 SDK 通用框架** | | | | | |
| 1.1.1 | 创建 `internal/platform/` 包目录 + `client.go` 定义 PlatformClient 接口 | 2 | — | internal/platform/client.go | 编译通过 |
| 1.1.2 | 定义 PlatformAccount、PlatformAuthToken、PublishRequest、PublishResponse 结构体 | 3 | 1.1.1 | types.go | 编译通过 |
| 1.1.3 | 实现 OAuth2.0 通用流程：AuthURL 生成 → Code 回调 → Token 交换 → Refresh | 6 | 1.1.2 | oauth.go | 单元测试覆盖 3 种场景 |
| 1.1.4 | 实现 API Key / Cookie 认证适配器（为不支持 OAuth 的平台） | 3 | 1.1.2 | auth.go | 单元测试 |
| 1.1.5 | 实现 Token 管理器：自动刷新、过期告警、持久化 | 4 | 1.1.3 | token_manager.go | Token 到期自动刷新 |
| 1.1.6 | 实现平台 SDK 注册表：map[string]PlatformClientFactory | 2 | 1.1.1 | registry.go | Register/Lookup 测试 |
| **1.2 抖音平台适配（首个完整平台）** | | | | | |
| 1.2.1 | 实现 DouyinClient：OAuth 授权流程（抖音开放平台） | 8 | 1.1.4 | douyin.go, douyin_auth.go | 回调 token 换取成功 |
| 1.2.2 | 实现 DouyinClient：视频发布 API（/video/create/） | 8 | 1.2.1 | douyin_publish.go | 模拟请求格式正确 |
| 1.2.3 | 实现 DouyinClient：图文发布 API（/note/create/） | 4 | 1.2.1 | douyin_note.go | 模拟请求格式正确 |
| 1.2.4 | 实现 DouyinClient：发布状态查询 + 错误处理 | 3 | 1.2.2 | douyin_status.go | 状态轮询正确 |
| 1.2.5 | 实现 DouyinClient：评论列表获取 + 回复 | 4 | 1.2.1 | douyin_comment.go | API 调用成功 |
| 1.2.6 | 编写抖音 SDK 单元测试（mock HTTP 服务器） | 4 | 1.2.2-1.2.5 | douyin_test.go | 全部通过 |
| **1.3 MCP Server 核心** | | | | | |
| 1.3.1 | 创建 `internal/mcp/` 包目录 + Server 核心结构体 | 2 | — | mcp/server.go | 编译通过 |
| 1.3.2 | 实现 SSE 传输（/sse 端点 + 消息流） | 6 | 1.3.1 | transport.go | curl SSE 端点返回 event stream |
| 1.3.3 | 实现 Tool 注册与路由：AgentPlugin → MCP Tool 映射 | 4 | 1.3.1 | tools.go | 注册 5 个现有 Agent 为 Tool |
| 1.3.4 | 实现 Resource 暴露（用量统计、发布状态） | 3 | 1.3.1 | resources.go | 读取资源返回正确数据 |
| 1.3.5 | 实现 MCP Server HTTP Handler（net/http 包装） | 3 | 1.3.2 | handler.go | ServeMux 路由正确 |
| 1.3.6 | config.yaml MCP 配置段：enabled / port / transport | 1 | 1.3.1 | config 扩展 | 配置可解析 |
| 1.3.7 | cmd/opc-agent MCP 启动逻辑 | 2 | 1.3.5 + 1.3.6 | main.go 扩展 | 服务启动 + 关闭 |
| **1.4 数据库扩展** | | | | | |
| 1.4.1 | 设计 platform_accounts 表 DDL + 迁移脚本 | 2 | — | migrations/001_platform_accounts.sql | psql 执行成功 |
| 1.4.2 | 实现 PlatformAccount 数据模型 + Repository | 4 | 1.4.1 | saas/platform_account.go | CRUD 单元测试 |
| 1.4.3 | Token 加密存储（AES-256-GCM 加密 AccessToken/RefreshToken） | 4 | 1.4.2 | token_crypto.go | 加密→解密一致性 |
| **1.5 平台账号管理 API** | | | | | |
| 1.5.1 | POST /api/v1/platform/accounts — 绑定账号（OAuth 发起 + callback） | 6 | 1.4.2 | handler_account.go | OAuth 流程完整 |
| 1.5.2 | GET /api/v1/platform/accounts — 账号列表（按平台筛选） | 3 | 1.4.2 | 同上 | 返回正确列表 |
| 1.5.3 | DELETE /api/v1/platform/accounts/:id — 解绑 | 2 | 1.4.2 | 同上 | 软删除成功 |
| 1.5.4 | PUT /api/v1/platform/accounts/:id — 编辑别名 | 2 | 1.4.2 | 同上 | 更新成功 |
| 1.5.5 | 平台账号绑定 UI 页面：OAuth 对接、账号列表、解绑确认弹窗 | 6 | 1.5.1-1.5.3 | web/src/pages/PlatformAccounts.jsx | 绑定流程完整 |

**WBS 1 合计：27 tasks / 104h**

---

### WBS 2 — Wave B2：Publish 引擎 + 内容发布全链路（Week 2-4）

**目标**：Publish Agent、发布引擎队列/重试/定时、内容适配、发布 UI

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **2.1 Publish Agent 插件** | | | | | |
| 2.1.1 | 注册 publisher AgentPlugin：Name / Info / Execute / Review | 3 | — | plugins/publisher/plugin.go | make run @publisher 响应 |
| 2.1.2 | 编写 publisher system_prompt（发布专家角色 + 各平台规则知识） | 3 | 2.1.1 | 同上 | Prompt 覆盖 13 平台规则 |
| 2.1.3 | 内容适配器 LLM 调用：根据目标平台自动转换文案/标签格式 | 6 | 2.1.1 | adapter.go | 同一内容输 3 平台输出不同 |
| 2.1.4 | 媒体文件处理：图片缩放/裁切、视频格式检测、自动转码提示 | 4 | 2.1.1 | media.go | 图片裁切比例正确 |
| 2.1.5 | 发布内容验证器：发布前检查必填字段、格式合规 | 3 | 2.1.1 | validator.go | 缺字段报错明确 |
| **2.2 发布引擎** | | | | | |
| 2.2.1 | 创建 `internal/publish/` 包 + Engine 核心结构体 | 2 | — | publish/engine.go | 编译通过 |
| 2.2.2 | 发布队列：内存队列 + 持久化（PostgreSQL） | 6 | 2.2.1 | queue.go | 队列 Push/Pop/Ack 正确 |
| 2.2.3 | 发布状态机：pending → publishing → success/failed | 4 | 2.2.1 | status.go | 状态转换完整测试 |
| 2.2.4 | 失败重试器：可配置最大重试次数 + 指数退避 | 4 | 2.2.2 | retry.go | 重试 3 次后最终失败 |
| 2.2.5 | 定时调度器：cron 表达式 → 定时触发发布 | 6 | 2.2.2 | scheduler.go | 定时任务准时触发 |
| 2.2.6 | 平台选择器：多平台并行发布 + 逐平台进度追踪 | 4 | 2.2.1 | dispatcher.go | 2 平台同时发布成功 |
| 2.2.7 | 发布任务日志：完整记录请求/响应/错误 | 3 | 2.2.3 | logger.go | 日志可追溯 |
| **2.3 数据库 + 数据模型** | | | | | |
| 2.3.1 | 设计 publish_tasks 表 DDL + 迁移脚本 | 2 | 1.4.1 | migrations/002_publish_tasks.sql | psql 执行成功 |
| 2.3.2 | 设计 content_items 素材表 + media_files 媒体文件表 | 2 | 2.3.1 | migrations/003_content.sql | 关联关系正确 |
| 2.3.3 | 实现 PublishTask Repository（CRUD + 状态查询） | 4 | 2.3.1 | saas/publish_task.go | 单元测试 |
| 2.3.4 | 实现 ContentItem Repository | 3 | 2.3.2 | saas/content_item.go | 单元测试 |
| **2.4 发布 API** | | | | | |
| 2.4.1 | POST /api/v1/publish — 创建发布任务 | 6 | 2.2.3 + 2.3.3 | handler_publish.go | 任务创建成功 |
| 2.4.2 | GET /api/v1/publish/tasks — 任务列表（按状态/平台/时间筛选） | 4 | 2.3.3 | 同上 | 列表 + 筛选正确 |
| 2.4.3 | GET /api/v1/publish/tasks/:id — 任务详情 + 进度 | 3 | 2.3.3 | 同上 | 详情完整 |
| 2.4.4 | POST /api/v1/publish/:id/cancel — 取消发布 | 2 | 2.2.3 | 同上 | 状态变更正确 |
| 2.4.5 | GET /api/v1/publish/calendar — 发布日历数据 | 3 | 2.3.3 | 同上 | 返回日历格式数据 |
| **2.5 发布 UI** | | | | | |
| 2.5.1 | 发布表单页面：标题/正文/媒体上传/平台选择/账号选择/定时 | 8 | 2.4.1 | PublishForm.jsx | 表单完整可提交 |
| 2.5.2 | 发布任务列表页：状态标签/进度条/操作按钮/筛选 | 6 | 2.4.2 | PublishList.jsx | 列表+筛选+操作 |
| 2.5.3 | 发布任务详情页：进度详情/各平台状态/日志/错误信息 | 5 | 2.4.3 | PublishDetail.jsx | 详情展示完整 |
| 2.5.4 | 发布日历页：月视图/周视图/拖拽排期/任务卡片 | 8 | 2.4.5 | PublishCalendar.jsx | 拖拽 + 视图切换 |
| 2.5.5 | 发布预览弹窗：各平台 Tab 切换效果预览 | 4 | 2.1.3 | PublishPreview.jsx | 3 平台预览不同 |
| 2.5.6 | 内容管理页：素材库上传/分类/搜索/选择 | 6 | 2.3.4 | ContentLibrary.jsx | 上传 → 搜索 → 选择 |

**WBS 2 合计：29 tasks / 118h**

---

### WBS 3 — Wave B3：Engage 引擎 + 互动运营（Week 4-6）

**目标**：Engage Agent、跨平台评论管理/自动回复、品牌监测、互动看板

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **3.1 Engage Agent 插件** | | | | | |
| 3.1.1 | 注册 engage AgentPlugin：Name / Info / Execute / Review | 3 | — | plugins/engage/plugin.go | make run @engage 响应 |
| 3.1.2 | 编写 engage system_prompt（互动运营专家角色 + 各平台互动规则） | 3 | 3.1.1 | 同上 | Prompt 覆盖回复/监测/趋势 |
| **3.2 评论管理核心** | | | | | |
| 3.2.1 | 创建 `internal/engage/` 包 + 评论采集器：各平台评论 API 轮询 | 8 | 1.2.5 + 其他平台 SDK | comment_collector.go | 跨平台评论聚合 |
| 3.2.2 | 评论聚合存储：统一评论模型 + 去重 + 增量同步 | 4 | 3.2.1 | comment_store.go | 同一评论不重复 |
| 3.2.3 | 评论情感分析：LLM 调用 → 正面/负面/中性 + 置信度 | 6 | 3.2.1 | sentiment.go | 情感分类准确 > 80% |
| 3.2.4 | 评论挖掘引擎：关键词识别 + 购买意向信号检测 + 商机打分 | 6 | 3.2.3 | mining.go | 标记「怎么买」等高意向 |
| **3.3 自动回复引擎** | | | | | |
| 3.3.1 | 自动回复规则引擎：关键词触发/全部回复/白名单/黑名单 | 6 | 3.2.1 | reply_rules.go | 规则匹配正确 |
| 3.3.2 | AI 回复生成器：LLM 根据评论 + 品牌语气生成个性化回复 | 8 | 3.3.1 | reply_generator.go | 回复不重复、语气一致 |
| 3.3.3 | 回复执行器：跨平台调用回复 API + 速率限制 + 错误处理 | 4 | 3.3.2 | reply_executor.go | 各平台回复成功 |
| 3.3.4 | 回复日志：自动回复记录 + 人工确认队列 | 3 | 3.3.3 | reply_log.go | 日志完整 |
| **3.4 品牌监测** | | | | | |
| 3.4.1 | 关键词监测引擎：配置关键词 → 跨平台搜索 → 结果聚合 | 6 | 3.2.1 | monitor.go | 品牌关键词匹配正确 |
| 3.4.2 | 实时告警：负面提及 → 邮件/站内通知 | 4 | 3.4.1 | alert.go | 负面内容触发告警 |
| 3.4.3 | 品牌监测看板：提及趋势图 / 情感分布 / 平台对比 | 4 | 3.4.1 | monitor_dashboard.go | 图表数据正确 |
| **3.5 数据库** | | | | | |
| 3.5.1 | 设计 comments 表 + reply_rules 表 DDL | 2 | — | migrations/004_engage.sql | psql 执行成功 |
| 3.5.2 | 设计 brand_mentions 表 + alerts 表 DDL | 2 | — | migrations/005_monitor.sql | psql 执行成功 |
| 3.5.3 | 实现 Comment Repository + ReplyRule Repository | 4 | 3.5.1 | saas/engage_repo.go | CRUD 测试 |
| **3.6 Engage API** | | | | | |
| 3.6.1 | GET /api/v1/engage/comments — 评论列表（平台/内容/时间筛选） | 4 | 3.5.3 | handler_engage.go | 列表 + 筛选正确 |
| 3.6.2 | POST /api/v1/engage/reply — 执行/计划自动回复 | 4 | 3.3.3 | 同上 | 回复成功 |
| 3.6.3 | GET /api/v1/engage/reply-rules — 回复规则列表 | 2 | 3.3.1 | 同上 | 规则完整 |
| 3.6.4 | PUT /api/v1/engage/reply-rules/:id — 编辑回复规则 | 2 | 3.3.1 | 同上 | 规则更新成功 |
| 3.6.5 | GET /api/v1/engage/stats — 互动数据汇总 | 4 | 3.2.1 | 同上 | 数据准确 |
| 3.6.6 | GET /api/v1/engage/monitor — 品牌监测数据 | 3 | 3.4.1 | 同上 | 监测数据完整 |
| **3.7 Engage UI** | | | | | |
| 3.7.1 | 评论管理中心页：跨平台评论流 / 筛选 / 回复 / 标注 | 8 | 3.6.1-3.6.2 | EngageComments.jsx | 评论加载 + 回复发送 |
| 3.7.2 | 自动回复规则配置页：规则 CRUD / 测试 / 白名单/黑名单 | 6 | 3.6.3-3.6.4 | ReplyRules.jsx | 规则增删改查 |
| 3.7.3 | 互动数据看板：总赞评藏/平台对比/趋势图/导出 | 6 | 3.6.5 | EngageDashboard.jsx | 图表数据正确 |
| 3.7.4 | 品牌监测页：提及列表 / 情感分布 / 告警配置 / 趋势 | 6 | 3.6.6 | BrandMonitor.jsx | 监测结果可看 |

**WBS 3 合计：28 tasks / 120h**

---

### WBS 4 — Wave B4：Create Agent 增强 + MCP 完善 + 其余平台 SDK（Week 6-8）

**目标**：AI 图片/视频生成、MCP 工具集完善、剩余平台 SDK、集成测试

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **4.1 Create Agent 图片生成** | | | | | |
| 4.1.1 | 创建 `internal/media/` 包 + ImageGenerator 接口 | 2 | — | media/image.go | 编译通过 |
| 4.1.2 | 实现 Midjourney API 适配器（imagine / vary / upscale） | 8 | 4.1.1 | midjourney.go | 模拟请求格式正确 |
| 4.1.3 | 实现 DALL-E 3 API 适配器 | 4 | 4.1.1 | dalle.go | 模拟请求格式正确 |
| 4.1.4 | 图片生成结果缓存 + 去重（相同 prompt 不重复请求） | 3 | 4.1.2 | image_cache.go | 重复 prompt 命中缓存 |
| 4.1.5 | 图片格式转换：生成结果 → 发布可用格式（裁切/压缩/水印） | 4 | 4.1.2 | image_processor.go | 输出合规 |
| **4.2 Create Agent 视频生成** | | | | | |
| 4.2.1 | 实现 Seedance API 适配器（text-to-video） | 8 | — | seedance.go | 模拟请求格式正确 |
| 4.2.2 | 实现 Kling API 适配器（text-to-video / image-to-video） | 8 | — | kling.go | 模拟请求格式正确 |
| 4.2.3 | 视频生成状态轮询 + 结果下载 | 4 | 4.2.1 | video_poller.go | 轮询至完成 |
| 4.2.4 | 文字转语音（TTS）集成：脚本配音自动生成 | 6 | — | tts.go | 配音时长匹配脚本 |
| 4.2.5 | 自动字幕生成：SRT 字幕文件 + 内嵌字幕 | 4 | 4.2.4 | subtitle.go | 时间轴对齐 |
| **4.3 Create Agent 插件注册** | | | | | |
| 4.3.1 | 注册 content_creator AgentPlugin（整合图片+视频+文案生成） | 3 | 4.1.2 + 4.2.2 | plugins/content_creator/plugin.go | @create 响应 |
| 4.3.2 | 批量生产模式：一个主题 → 多个变体（不同平台/风格/尺寸） | 6 | 4.3.1 | batch.go | 5 变体同时生成 |
| 4.3.3 | 内容扩写/缩写：短内容→长文 / 长文→摘要 | 4 | 4.3.1 | rewrite.go | 输入输出比例符合 |
| **4.4 剩余平台 SDK（Wave 1 未覆盖）** | | | | | |
| 4.4.1 | 小红书 SDK（XiaohongshuClient）：OAuth + 图文发布 + 笔记管理 | 8 | 1.1.4 | xiaohongshu.go | API 调用成功 |
| 4.4.2 | B站 SDK（BilibiliClient）：OAuth + 视频发布 + 动态发布 | 8 | 1.1.4 | bilibili.go | API 调用成功 |
| 4.4.3 | 快手 SDK（KuaishouClient）：OAuth + 视频发布 | 8 | 1.1.4 | kuaishou.go | API 调用成功 |
| 4.4.4 | 微信 SDK（WeChatClient）：视频号 + 公众号接入 | 8 | 1.1.4 | wechat.go | API 调用成功 |
| 4.4.5 | YouTube SDK：YouTube Data API v3（OAuth + 视频上传） | 8 | 1.1.4 | youtube.go | API 调用成功 |
| 4.4.6 | TikTok SDK：TikTok Business API（OAuth + 视频发布） | 8 | 1.1.4 | tiktok.go | API 调用成功 |
| 4.4.7 | Facebook/Instagram SDK：Graph API 封装 | 6 | 1.1.4 | facebook.go | API 调用成功 |
| 4.4.8 | Twitter(X) SDK：Twitter API v2 | 4 | 1.1.4 | twitter.go | API 调用成功 |
| 4.4.9 | Pinterest + LinkedIn + Threads SDK（V2+ 可选平台） | 6 | 1.1.4 | pinterest.go, linkedin.go, threads.go | API 调用成功 |
| **4.5 MCP 完善** | | | | | |
| 4.5.1 | 将 publisher Agent 暴露为 MCP Tool | 2 | 1.3.3 + 2.1.1 | mcp/tools.go 扩展 | Claude 可调用 |
| 4.5.2 | 将 engage Agent 暴露为 MCP Tool | 2 | 1.3.3 + 3.1.1 | 同上 | Claude 可调用 |
| 4.5.3 | 将 content_creator Agent 暴露为 MCP Tool | 2 | 1.3.3 + 4.3.1 | 同上 | Claude 可调用 |
| 4.5.4 | 工作流暴露为 MCP Tool：YAML 工作流自动注册 | 4 | 1.3.3 | workflow_tool.go | 工作流可触发 |
| 4.5.5 | MCP 客户端 SDK 示例（Node.js + Python） | 4 | 1.3.5 | examples/mcp-client.js, .py | 连接并调用成功 |
| **4.6 内容适配增强（跨平台）** | | | | | |
| 4.6.1 | 视频比例自动适配：16:9 ↔ 9:16 ↔ 1:1 智能裁切提示 | 4 | 2.1.3 | aspect_ratio.go | 裁切区域智能选择 |
| 4.6.2 | 文案长度自适应：平台字符限制检测 + 自动截断/扩写 | 3 | 2.1.3 | length_adapter.go | 超长文案自动处理 |
| 4.6.3 | 标签格式转换：不同平台标签规则（#tag / @tag / [tag]） | 2 | 2.1.3 | tag_adapter.go | 格式正确 |
| **4.7 集成测试 + 文档** | | | | | |
| 4.7.1 | Phase B 全链路集成测试（账号绑定→创建→发布→互动） | 8 | 全部 WBS 1-4 | tests/integration/phase_b_test.go | 全场景通过 |
| 4.7.2 | Phase B 性能测试（并发发布 50 任务） | 4 | 4.7.1 | tests/bench/phase_b_bench.go | P50 < 2s, P99 < 10s |
| 4.7.3 | Phase B 用户手册 + API 文档 | 6 | 4.7.1 | docs/phase-b-manual.md | 覆盖所有功能 |

**WBS 4 合计：29 tasks / 148h**

---

## Phase A — 内容创作增强层（WBS 5-7）

总计：**3 个 Wave，37 项任务，3 周**

---

### WBS 5 — Wave A1：video_script Agent（Week 9）

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **5.1 video_script Agent 插件** | | | | | |
| 5.1.1 | 注册 video_script AgentPlugin | 3 | — | plugins/video_script/plugin.go | @script 响应 |
| 5.1.2 | 编写 video_script system_prompt（脚本创作专家 + 7 平台风格知识） | 4 | 5.1.1 | 同上 | Prompt 覆盖全平台 |
| 5.1.3 | 实现脚本生成核心逻辑：topic → LLM → 结构化脚本(分镜/话术/画面) | 8 | 5.1.1 | script_generator.go | 输出符合 JSON Schema |
| 5.1.4 | 实现平台风格适配器：7 平台（抖音/快手/B站/TikTok/YouTube/小红书/微信）时长/比例/话术 | 6 | 5.1.3 | platform_adapter.go | 各平台输出风格差异明显 |
| 5.1.5 | 实现话术风格模板：7 种风格（带货/知识/娱乐/教程/测评/Vlog/剧情） | 4 | 5.1.3 | style_templates.go | 各风格输出符合调性 |
| 5.1.6 | 实现批量生成：同一主题 → 最多 5 个不同角度脚本变体 | 4 | 5.1.3 | batch_gen.go | 5 变体不重复 |
| 5.1.7 | 脚本输出验证器：时长校验 + 分镜完整性 + 必填字段 | 3 | 5.1.3 | script_validator.go | 缺字段报错 |
| **5.2 video_script API** | | | | | |
| 5.2.1 | POST /api/v1/create/video-script — 生成视频脚本 | 4 | 5.1.6 | handler_video_script.go | 返回完整脚本 |
| **5.3 video_script UI** | | | | | |
| 5.3.1 | 脚本生成面板：主题/平台/风格/时长配置 → 生成 | 6 | 5.2.1 | VideoScriptPanel.jsx | 完整配置→生成 |
| 5.3.2 | 脚本预览组件：分镜时间线视图 / 文案视图 / 分镜详情卡片 | 8 | 5.3.1 | ScriptPreview.jsx | 时间线可拖动 |
| 5.3.3 | 脚本导出：复制文案 / 导出 JSON / 导出模板 | 3 | 5.3.2 | ScriptExport.jsx | 格式正确 |
| 5.3.4 | 一键跳转发布：脚本 → Publish Agent 预填内容 | 3 | 5.3.1 + 2.5.1 | ScriptToPublish.jsx | 跳转携带脚本内容 |

**WBS 5 合计：12 tasks / 56h**

---

### WBS 6 — Wave A2：tag_generator Agent（Week 10）

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **6.1 tag_generator Agent 插件** | | | | | |
| 6.1.1 | 注册 tag_generator AgentPlugin | 2 | — | plugins/tag_generator/plugin.go | @tags 响应 |
| 6.1.2 | 编写 tag_generator system_prompt（标签优化专家 + 各平台标签策略） | 3 | 6.1.1 | 同上 | Prompt 覆盖全平台策略 |
| 6.1.3 | 标签生成核心：内容分析 → 标签提取 → 热度评估 → 分层输出 | 6 | 6.1.1 | tag_generator.go | 输出符合 JSON Schema |
| 6.1.4 | 标签热度模型：流量标签(🔥火爆) / 上升标签(📈上升) / 长尾标签(💧长尾) | 4 | 6.1.3 | heat_model.go | 热度分级合理 |
| 6.1.5 | 标签策略引擎：大标签+小标签组合 / 流量+精准搭配 / 平台差异化 | 4 | 6.1.3 | strategy.go | 策略建议可操作 |
| 6.1.6 | 标签去重 + 冲突检测 | 2 | 6.1.3 | dedup.go | 无重复标签 |
| **6.2 tag_generator API** | | | | | |
| 6.2.1 | POST /api/v1/create/tags — 生成标签推荐 | 3 | 6.1.5 | handler_tags.go | 返回完整标签列表 |
| **6.3 tag_generator UI** | | | | | |
| 6.3.1 | 标签生成面板：内容输入 + 平台选择 + 分类选择 → 生成 | 4 | 6.2.1 | TagGeneratorPanel.jsx | 完整配置→生成 |
| 6.3.2 | 标签云图组件：按热度大小排列 / 分层展示 / 点击复制 | 4 | 6.3.1 | TagCloud.jsx | 热度大小映射正确 |
| 6.3.3 | 标签策略卡片：推荐组合 / 优化建议 / 不推荐标签 | 3 | 6.3.1 | TagStrategy.jsx | 建议清晰可读 |

**WBS 6 合计：10 tasks / 35h**

---

### WBS 7 — Wave A3：trend_radar Agent（Week 10-11）

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **7.1 热点数据采集** | | | | | |
| 7.1.1 | 抖音热榜采集器：Web Scraping / 第三方 API → 结构化热榜数据 | 6 | — | collector_douyin.go | 数据格式正确 |
| 7.1.2 | 小红书热搜采集器 | 4 | — | collector_xiaohongshu.go | 数据格式正确 |
| 7.1.3 | B站热门采集器 | 4 | — | collector_bilibili.go | 数据格式正确 |
| 7.1.4 | 微博热搜采集器 | 3 | — | collector_weibo.go | 数据格式正确 |
| 7.1.5 | 快手热榜采集器 | 3 | — | collector_kuaishou.go | 数据格式正确 |
| 7.1.6 | TikTok Trends 采集器 | 4 | — | collector_tiktok.go | 数据格式正确 |
| 7.1.7 | YouTube Trends 采集器（YouTube API） | 4 | — | collector_youtube.go | 数据格式正确 |
| 7.1.8 | 定时采集调度器：可配置频率 + 数据去重 + 增量存储 | 6 | 7.1.1-7.1.7 | collector_scheduler.go | 定时触发采集 |
| **7.2 trend_radar Agent** | | | | | |
| 7.2.1 | 注册 trend_radar AgentPlugin | 2 | — | plugins/trend_radar/plugin.go | @trend 响应 |
| 7.2.2 | 编写 trend_radar system_prompt（趋势分析专家） | 3 | 7.2.1 | 同上 | Prompt 覆盖分析维度 |
| 7.2.3 | 热点分析引擎：话题热度计算 / 趋势方向判定 / 生命周期评估 | 8 | 7.1.8 | trend_analyzer.go | 趋势判定准确 > 80% |
| 7.2.4 | 选题推荐：基于账号定位 + 热度 + 相关性 + 竞争度 → 推荐排序 | 6 | 7.2.3 | recommendation.go | 推荐合理可解释 |
| 7.2.5 | 竞品监测模式：竞品账号内容分析 → 热门话题分布 | 4 | 7.2.3 | competitor.go | 竞品话题提取正确 |
| 7.2.6 | 日报/周报生成器：LLM 生成报告 → 结构化 Markdown | 6 | 7.2.3 | report_gen.go | 报告完整可读 |
| **7.3 数据库** | | | | | |
| 7.3.1 | 设计 hot_topics 表 + trend_reports 表 DDL | 2 | — | migrations/006_trend.sql | psql 执行成功 |
| 7.3.2 | 实现 HotTopic Repository | 3 | 7.3.1 | saas/trend_repo.go | CRUD 测试 |
| **7.4 trend_radar API** | | | | | |
| 7.4.1 | POST /api/v1/create/trend-report — 生成热点报告 | 4 | 7.2.4 | handler_trend.go | 报告完整 |
| 7.4.2 | GET /api/v1/create/trend/hotlist — 当前热榜数据 | 3 | 7.1.8 + 7.3.2 | 同上 | 热榜数据最新 |
| **7.5 trend_radar UI** | | | | | |
| 7.5.1 | 热点趋势看板：平台选择 → 热榜列表 + 趋势图表 + 词云 | 8 | 7.4.1-7.4.2 | TrendDashboard.jsx | 数据可视化正确 |
| 7.5.2 | 选题推荐卡片：话题 + 平台 + 置信度 + 推荐理由 + 「采纳创作」 | 6 | 7.4.1 | TrendRecommendations.jsx | 推荐可操作 |
| 7.5.3 | 热点报告页：日报/周报切换 / 可视化 / 导出 PDF | 6 | 7.2.6 | TrendReport.jsx | 报告可导出 |

**WBS 7 合计：15 tasks / 89h**

---

## Phase C — 商业化变现层（WBS 8-11）

总计：**4 个 Wave，90 项任务，16 周**

---

### WBS 8 — Wave C1：Marketplace 核心 + 数据库 + API（Week 12-15）

**目标**：任务发布/浏览/接单核心功能、任务 CRUD、商家和创作者角色

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **8.1 数据库设计** | | | | | |
| 8.1.1 | 设计 marketplace_tasks 表 DDL（含 settlement_type, budget, requirements 等） | 3 | — | migrations/007_marketplace_tasks.sql | psql 执行成功 |
| 8.1.2 | 设计 task_orders 表 DDL（含接单/交付/审核状态机） | 3 | 8.1.1 | migrations/008_task_orders.sql | 状态约束完整 |
| 8.1.3 | 设计 wallets 表 + wallet_transactions 表 DDL | 3 | — | migrations/009_wallet.sql | psql 执行成功 |
| 8.1.4 | 设计 reviews 表 DDL（含多维度评分 JSON） | 2 | — | migrations/010_reviews.sql | psql 执行成功 |
| 8.1.5 | 所有表添加外键约束 + 索引 + RLS 策略（多租户隔离） | 4 | 8.1.1-8.1.4 | migrations/011_constraints.sql | 约束检查通过 |
| **8.2 数据模型 + Repository** | | | | | |
| 8.2.1 | 实现 MarketplaceTask 模型 + Repository（CRUD + 筛选/分页） | 8 | 8.1.1 | saas/marketplace_task.go | 单元测试 |
| 8.2.2 | 实现 TaskOrder 模型 + Repository（状态机转换） | 6 | 8.1.2 | saas/task_order.go | 状态转换合法 |
| 8.2.3 | 实现 Wallet 模型 + Repository | 4 | 8.1.3 | saas/wallet.go | 余额计算正确 |
| 8.2.4 | 实现 WalletTransaction 模型 + Repository | 4 | 8.1.3 | saas/transaction.go | 流水完整 |
| 8.2.5 | 实现 Review 模型 + Repository | 3 | 8.1.4 | saas/review.go | CRUD 测试 |
| **8.3 用户角色系统** | | | | | |
| 8.3.1 | 扩展 User 模型：添加 role 字段（advertiser / creator / both） | 2 | — | saas/user.go 扩展 | 角色类型正确 |
| 8.3.2 | SaaS Auth 适配：角色权限中间件 | 4 | 8.3.1 | middleware.go | 角色路由拦截正确 |
| **8.4 Marketplace API（商家侧）** | | | | | |
| 8.4.1 | POST /api/v1/marketplace/tasks — 发布任务 | 6 | 8.2.1 + 8.3.2 | handler_marketplace.go | 任务创建成功 |
| 8.4.2 | PUT /api/v1/marketplace/tasks/:id — 编辑任务 | 3 | 8.4.1 | 同上 | 更新成功 |
| 8.4.3 | GET /api/v1/marketplace/tasks — 我的任务列表 | 4 | 8.2.1 | 同上 | 列表正确 |
| 8.4.4 | GET /api/v1/marketplace/tasks/:id — 任务详情 | 2 | 8.2.1 | 同上 | 详情完整 |
| 8.4.5 | POST /api/v1/marketplace/tasks/:id/cancel — 取消任务 | 2 | 8.4.1 | 同上 | 状态变更 |
| **8.5 Marketplace API（创作者侧）** | | | | | |
| 8.5.1 | GET /api/v1/marketplace/open-tasks — 开放任务列表（筛选/分页） | 4 | 8.2.1 | handler_marketplace_open.go | 筛选条件正确 |
| 8.5.2 | GET /api/v1/marketplace/open-tasks/:id — 开放任务详情 | 2 | 8.2.1 | 同上 | 详情不含敏感字段 |
| 8.5.3 | POST /api/v1/marketplace/tasks/:id/apply — 接单（含资格检查） | 6 | 8.2.2 + 8.2.1 | handler_marketplace_order.go | 资格检查正确 |
| 8.5.4 | GET /api/v1/marketplace/my-orders — 我的接单列表 | 3 | 8.2.2 | 同上 | 列表正确 |
| 8.5.5 | GET /api/v1/marketplace/orders/:id — 订单详情 | 2 | 8.2.2 | 同上 | 详情完整 |
| **8.6 Marketplace UI（商家侧）** | | | | | |
| 8.6.1 | 任务发布表单页：标题/描述/平台/结算方式/预算/要求/截止日期 | 8 | 8.4.1 | TaskCreateForm.jsx | 所有字段可填可提交 |
| 8.6.2 | 模板任务选择器：预设模板 → 自动填充表单 | 4 | 8.6.1 | TaskTemplateSelector.jsx | 模板填充正确 |
| 8.6.3 | 任务管理页：我发布的任务列表/状态筛选/操作 | 6 | 8.4.3 | MyTasksList.jsx | 列表+筛选+操作 |
| 8.6.4 | 任务详情页（商家视角）：接单创作者列表/数据总览 | 6 | 8.4.4 | TaskDetailAdvertiser.jsx | 创作者列表正确 |
| **8.7 Marketplace UI（创作者侧）** | | | | | |
| 8.7.1 | 任务广场页：筛选（平台/结算/预算/领域）+ 任务卡片 + 智能推荐 | 8 | 8.5.1 | TaskMarketplace.jsx | 筛选+推荐+浏览 |
| 8.7.2 | 任务详情页（创作者视角）：要求/预算/截止 → 接单按钮 | 4 | 8.5.2 | TaskDetailCreator.jsx | 资格提示正确 |
| 8.7.3 | 我的订单页：进行中/已完成/已取消订单列表 | 6 | 8.5.4 | MyOrdersList.jsx | 订单状态正确 |

**WBS 8 合计：27 tasks / 118h**

---

### WBS 9 — Wave C2：订单流 + 内容交付 + 审核（Week 16-19）

**目标**：交付/审核/修改循环、定向邀请、智能推荐

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **9.1 内容交付流程** | | | | | |
| 9.1.1 | POST /api/v1/marketplace/orders/:id/deliver — 交付内容（链接+截图） | 6 | 8.5.4 | handler_deliver.go | 交付成功 |
| 9.1.2 | 交付物存储：图片/PDF 文件上传 + 链接验证 | 4 | 9.1.1 | deliver_store.go | 上传+链接验证 |
| 9.1.3 | 交付状态机：delivered → approved(结算) / rejected(修改) | 4 | 9.1.1 | deliver_status.go | 状态转换正确 |
| 9.1.4 | 审核通知：商家收到审核提醒 / 创作者收到结果通知 | 4 | 9.1.3 | deliver_notify.go | 通知发送成功 |
| 9.1.5 | 交付超时处理：截止日期自动标记 | 3 | 9.1.3 | deliver_timeout.go | 超时自动状态变更 |
| **9.2 商家审核 API** | | | | | |
| 9.2.1 | POST /api/v1/marketplace/orders/:id/approve — 批准交付 | 3 | 9.1.3 | handler_review.go | 状态→approved |
| 9.2.2 | POST /api/v1/marketplace/orders/:id/reject — 驳回（含修改意见） | 3 | 9.1.3 | 同上 | 状态→rejected |
| 9.2.3 | GET /api/v1/marketplace/orders/:id/deliveries — 交付历史 | 2 | 9.1.1 | 同上 | 历史完整 |
| **9.3 定向邀请** | | | | | |
| 9.3.1 | POST /api/v1/marketplace/tasks/:id/invite — 邀请创作者 | 4 | 8.2.1 | handler_invite.go | 邀请发送成功 |
| 9.3.2 | GET /api/v1/marketplace/invitations — 我收到的邀请列表 | 3 | 9.3.1 | 同上 | 邀请列表正确 |
| 9.3.3 | POST /api/v1/marketplace/invitations/:id/accept — 接受邀请 | 2 | 9.3.1 | 同上 | 自动创建 order |
| **9.4 智能推荐** | | | | | |
| 9.4.1 | 创作者-任务匹配引擎：粉丝量/领域/互动率 → 匹配度评分 | 8 | 8.2.1 | recommendation_engine.go | 推荐匹配度 > 70% |
| 9.4.2 | 智能推荐 API：GET /api/v1/marketplace/recommendations | 4 | 9.4.1 | handler_recommend.go | 匹配任务列表 |
| 9.4.3 | 推荐理由生成：LLM 解释每个推荐的原因 | 4 | 9.4.2 | recommendation_reason.go | 理由可理解 |
| **9.5 UI** | | | | | |
| 9.5.1 | 交付提交表单：链接输入 + 截图上传 + 备注 | 6 | 9.1.1 | DeliverForm.jsx | 提价成功 |
| 9.5.2 | 审核面板（商家）：交付内容预览 + 批准/驳回/修改意见输入 | 6 | 9.2.1-9.2.2 | ReviewPanel.jsx | 审核操作正确 |
| 9.5.3 | 交付状态追踪组件：时间线视图 | 4 | 9.1.3 | DeliveryTimeline.jsx | 状态可视化正确 |
| 9.5.4 | 邀请弹窗：搜索创作者 + 选择 + 发送邀请 | 4 | 9.3.1 | InviteDialog.jsx | 邀请流程完整 |
| 9.5.5 | 智能推荐卡片：推荐理由 + 匹配度 + 采纳按钮 | 4 | 9.4.2 | RecommendationCard.jsx | 推荐可视化 |

**WBS 9 合计：24 tasks / 98h**

---

### WBS 10 — Wave C3：结算引擎 + 钱包 + 提现（Week 20-23）

**目标**：CPS/CPE/CPM 三种结算、自动/手动结算、钱包、提现

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **10.1 结算引擎** | | | | | |
| 10.1.1 | 创建 `internal/settlement/` 包 + SettlementEngine 接口 | 2 | — | settlement/engine.go | 编译通过 |
| 10.1.2 | CPS 结算器：成交额 × CPS 费率 = 结算金额 | 6 | 10.1.1 | cps.go | 计算精确到分 |
| 10.1.3 | CPE 结算器：互动数 × 加权单价（点赞/评论/收藏/分享不同权重） | 6 | 10.1.1 | cpe.go | 加权计算正确 |
| 10.1.4 | CPM 结算器：曝光量 ÷ 1000 × CPM 单价 | 4 | 10.1.1 | cpm.go | 千次曝光计算正确 |
| 10.1.5 | 混合结算器：保底 CPM + 激励 CPS | 6 | 10.1.2 + 10.1.4 | hybrid.go | 混合计算正确 |
| 10.1.6 | 平台佣金计算器：分级费率（按累计收益阶梯） | 4 | 10.1.1 | commission.go | 费率阶梯正确 |
| 10.1.7 | 结算单生成器：PDF 结算单 + 明细列表 | 6 | 10.1.2-10.1.4 | invoice.go | 结算单完整 |
| 10.1.8 | 结算状态机：pending → calculating → confirmed → paid | 4 | 10.1.1 | settlement_status.go | 状态转换合法 |
| **10.2 结算 API** | | | | | |
| 10.2.1 | POST /api/v1/settlement/calculate/:order_id — 手动触发结算 | 4 | 10.1.8 | handler_settlement.go | 结算金额正确 |
| 10.2.2 | GET /api/v1/settlement/orders — 结算列表 | 3 | 10.1.8 | 同上 | 列表正确 |
| 10.2.3 | GET /api/v1/settlement/orders/:id — 结算详情 + 明细 | 2 | 10.1.8 | 同上 | 明细完整 |
| 10.2.4 | GET /api/v1/settlement/invoices/:id — 下载结算单 PDF | 3 | 10.1.7 | 同上 | PDF 可下载 |
| **10.3 钱包系统** | | | | | |
| 10.3.1 | 钱包充值（商家）：POST /api/v1/wallet/deposit | 4 | 8.2.3 | handler_wallet.go | 充值成功 |
| 10.3.2 | 结算入账：审核通过 → 自动入账创作者钱包 | 6 | 10.1.8 + 8.2.3 | wallet_income.go | 入账金额正确 |
| 10.3.3 | 冻结/解冻：接单锁定预算 / 结算后解冻 | 4 | 10.3.2 | wallet_freeze.go | 冻结金额正确 |
| 10.3.4 | 提现申请：POST /api/v1/wallet/withdraw | 6 | 8.2.4 | handler_withdraw.go | 提现申请成功 |
| 10.3.5 | 提现审核（运营）：GET/POST withdraws 列表 + 审核 | 4 | 10.3.4 | handler_withdraw_admin.go | 审核流程完整 |
| 10.3.6 | 提现处理：对接支付宝/微信支付企业付款 API | 8 | 10.3.4 | withdraw_payment.go | 付款成功 |
| 10.3.7 | 收款方式绑定：支付宝/微信/银行卡信息存储（加密） | 4 | — | payment_method.go | 绑定+解密正确 |
| **10.4 钱包 API** | | | | | |
| 10.4.1 | GET /api/v1/wallet/balance — 查询余额 | 2 | 8.2.3 | handler_wallet_balance.go | 余额正确 |
| 10.4.2 | GET /api/v1/wallet/transactions — 交易流水 | 4 | 8.2.4 | 同上 | 流水完整 |
| 10.4.3 | GET /api/v1/wallet/withdraw-records — 提现记录 | 2 | 10.3.4 | 同上 | 记录完整 |
| **10.5 UI** | | | | | |
| 10.5.1 | 钱包页面：余额卡片（总余额/冻结/可提现/累计）+ 操作按钮 | 6 | 10.4.1-10.4.3 | WalletPage.jsx | 余额显示正确 |
| 10.5.2 | 提现弹窗：金额输入 + 收款方式选择 + 确认 | 4 | 10.3.4 | WithdrawDialog.jsx | 提现申请成功 |
| 10.5.3 | 交易流水表：时间/类型/金额/状态 + 筛选 | 4 | 10.4.2 | TransactionList.jsx | 流水列表正确 |
| 10.5.4 | 结算明细页：任务/订单/金额/佣金/结算单下载 | 4 | 10.2.3-10.2.4 | SettlementDetail.jsx | 明细完整 |

**WBS 10 合计：24 tasks / 106h**

---

### WBS 11 — Wave C4：评价体系 + 数据报告 + 反作弊 + 集成（Week 24-27）

**目标**：双端互评、信用分、数据分析、反作弊、管理员后台、全链路集成

| WBS | 任务 | 工时(h) | 前置 | 产出物 | 验证方式 |
|-----|------|---------|------|--------|----------|
| **11.1 评价体系** | | | | | |
| 11.1.1 | POST /api/v1/reviews — 创建评价（商家→创作者 / 创作者→商家） | 4 | 8.2.5 | handler_review.go | 评价创建成功 |
| 11.1.2 | GET /api/v1/reviews — 评价列表（按用户/任务筛选） | 3 | 8.2.5 | 同上 | 列表正确 |
| 11.1.3 | 评价统计：平均分 / 各维度分 / 评价数 | 3 | 11.1.1 | review_stats.go | 统计数据准确 |
| **11.2 信用分系统** | | | | | |
| 11.2.1 | 信用分计算引擎：各种行为 +- 分规则实现 | 6 | — | credit_score.go | 分数变化正确 |
| 11.2.2 | 信用分限制：低于 60 分限制接单/发单 | 4 | 11.2.1 | credit_restrict.go | 限制生效 |
| 11.2.3 | GET /api/v1/wallet/credit-score — 信用分查询 | 2 | 11.2.1 | handler_credit.go | 信用分正确 |
| **11.3 数据报告** | | | | | |
| 11.3.1 | 商家数据报告：任务总览/ROI/创作者对比/趋势（聚合查询） | 8 | 8.4.3 + 8.5.4 | report_advertiser.go | 数据准确 |
| 11.3.2 | 创作者数据报告：收益总览/接单分析/各平台分布 | 6 | 8.5.4 + 10.4.2 | report_creator.go | 数据准确 |
| 11.3.3 | 报告导出：Excel / PDF 导出 | 6 | 11.3.1-11.3.2 | report_export.go | 导出格式正确 |
| **11.4 反作弊系统** | | | | | |
| 11.4.1 | 刷量检测：互动数据异常检测（暴增/机器人模式识别） | 8 | — | anti_fraud_detect.go | 识别率 > 90% |
| 11.4.2 | 重复账号检测：同一设备/IP 多账号检测 | 4 | — | anti_fraud_account.go | 关联账号识别 |
| 11.4.3 | 内容查重：已发布内容相似度检测 | 4 | — | anti_fraud_content.go | 重复内容标记 |
| 11.4.4 | 风控告警：异常行为实时告警 + 自动冻结 | 4 | 11.4.1-11.4.3 | anti_fraud_alert.go | 告警触发正确 |
| **11.5 管理员后台** | | | | | |
| 11.5.1 | 任务管理（管理员）：查看/下架/调解争议 | 6 | 8.4.2 | admin_tasks.jsx + handler | 管理操作正确 |
| 11.5.2 | 用户管理（管理员）：查看/冻结/解冻/角色变更 | 4 | 8.3.1 | admin_users.jsx + handler | 管理操作正确 |
| 11.5.3 | 结算管理（管理员）：审核提现 / 手动结算 / 退款 | 6 | 10.2.1 + 10.3.5 | admin_settlement.jsx + handler | 管理操作正确 |
| 11.5.4 | 数据仪表盘（管理员）：平台总交易额/活跃用户/佣金收入 | 6 | 11.3.1 | admin_dashboard.jsx + handler | 数据准确 |
| **11.6 全链路集成测试 + 文档** | | | | | |
| 11.6.1 | Phase C 集成测试（任务发布→接单→交付→审核→结算→提现） | 12 | 全部 WBS 8-11 | tests/integration/phase_c_test.go | 全场景通过 |
| 11.6.2 | Phase C 性能测试（100 并发接单） | 4 | 11.6.1 | tests/bench/phase_c_bench.go | P50 < 1s |
| 11.6.3 | Phase C 用户手册 + API 文档 | 8 | 11.6.1 | docs/phase-c-manual.md | 覆盖所有功能 |
| 11.6.4 | 全平台回归测试（Phase B + A + C 全链路） | 8 | 4.7.1 + 11.6.1 | tests/e2e/full_chain_test.go | 全部通过 |

**WBS 11 合计：17 tasks / 110h**

---

## 汇总统计

### 按 Phase

| Phase | Waves | 任务数 | 总工时(h) | 预计周数 | 单人周数 |
|-------|-------|--------|----------|---------|---------|
| B | WBS 1-4 | 113 | 490 | 8 | 12.3 |
| A | WBS 5-7 | 37 | 180 | 3 | 4.5 |
| C | WBS 8-11 | 90 | 432 | 16 | 10.8 |
| **总计** | **11** | **240** | **1102** | **27** | **27.6** |

### 按职能

| 职能 | 工时(h) | 占比 |
|------|---------|------|
| Go 后端（API + 引擎 + SDK） | 510 | 46% |
| 前端 UI（React + Tailwind） | 310 | 28% |
| 数据库 + 数据模型 | 82 | 8% |
| 测试（单元 + 集成 + 性能） | 110 | 10% |
| 文档 + 运维 | 90 | 8% |
| **总计** | **1102** | **100%** |

### 并行度分析

| 波次 | 可并行最大人数 | 建议团队 |
|------|--------------|---------|
| WBS 1 (B1) | 3-4 | 2 backend + 1 DB + 1 MCP |
| WBS 2 (B2) | 2-3 | 1 backend + 1 frontend + 1 fullstack |
| WBS 3 (B3) | 2-3 | 1 backend + 1 frontend + 1 fullstack |
| WBS 4 (B4) | 4-5 | 2 SDK + 1 media + 1 MCP + 1 test |
| WBS 5-7 (A) | 2-3 | 1 backend + 1 frontend + 1 AI prompt |
| WBS 8 (C1) | 3-4 | 2 backend + 1 frontend + 1 DB |
| WBS 9 (C2) | 2-3 | 1 backend + 1 frontend + 1 fullstack |
| WBS 10 (C3) | 2-3 | 2 backend (settlement + wallet) + 1 frontend |
| WBS 11 (C4) | 3-4 | 1 fullstack + 1 anti-fraud + 1 admin + 1 test |

---

## 关键里程碑

| 里程碑 | 时间 | 交付物 | 验收标准 |
|--------|------|--------|----------|
| M1: 平台 SDK 可用 | Week 2 | 抖音 SDK + OAuth + Token 管理 | 绑定抖音账号 → 发布视频成功 |
| M2: Publish MVP | Week 4 | Publish Agent + 发布队列 + 内容适配 | 一次内容 → 抖音+小红书同时发布 |
| M3: Engage MVP | Week 6 | 评论管理 + 自动回复 + 互动看板 | 跨平台评论聚合 → AI 自动回复 |
| M4: Phase B 完成 | Week 8 | 全部 4 Wave | 全链路集成测试通过 |
| M5: Phase A 完成 | Week 11 | 3 个创作 Agent | 热点→脚本→标签全自动流水线 |
| M6: Marketplace MVP | Week 15 | 任务发布 + 接单 + 审核 | 商家创建任务 → 创作者交付 |
| M7: 结算引擎 | Week 23 | CPS/CPE/CPM 结算 | 结算单金额 100% 准确 |
| M8: Phase C 完成 | Week 27 | 全功能交易市场 | 全链路 e2e 测试通过 |
