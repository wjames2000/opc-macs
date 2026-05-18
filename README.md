# OPC-Agent

多任务智能协助系统 — 为 OPC（一人公司）创业者打造的 AI 内容营销全链路平台。

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://golang.org)
[![Wails](https://img.shields.io/badge/Wails-v2-DF0000?logo=wails)](https://wails.io)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

---

## 功能概览

```
✅ CLI REPL 交互              ✅ 5 个 Agent 插件
✅ 6 个模型 Provider 支持      ✅ 多轮对话 / 会话管理
✅ A2A Agent 链式调用          ✅ Token 统计 / 费用估算
✅ 工作流编排引擎              ✅ 结果缓存 (LRU+TTL)
✅ Docker 容器化               ✅ 配置热重载
✅ SaaS 多租户后端             ✅ React Web UI
✅ pgvector 向量存储           ✅ Wails 桌面端 (macOS)
✅ Chrome 扩展 (MV3)          ✅ 健康检查 API
✅ 插件脚手架                  ✅ MCP Protocol 支持
✅ 多平台内容发布 (12 平台)    ✅ 自动互动运营 (评论/监测/雷达)
✅ 视频脚本生成 (7 风格)       ✅ 热点雷达 (6 平台采集)
✅ AI 标签推荐 (热度/策略)     ✅ 内容交易市场 (接单/结算/钱包)
✅ 图片/视频 AI 生成           ✅ MCP 协议支持
```

---

## 项目阶段

OPC-Agent 内容营销升级已完成 **238/240 项任务 (~99%)**：

| Phase | 内容 | 状态 | 产出物 |
|-------|------|------|--------|
| B — 内容平台核心层 | 12 平台 SDK + 发布引擎 + 互动引擎 + MCP | 🟢 **100%** | `internal/platform/` `internal/publish/` `internal/engage/` |
| A — 内容创作增强层 | 视频脚本 + 标签推荐 + 热点雷达 Agent | 🟢 **100%** | `internal/plugins/video_script.go` `tag_generator.go` `trend_radar.go` |
| C — 商业化变现层 | 任务市场 + 结算引擎 + 钱包 + 反作弊 | 🟢 **~98%** | `internal/saas/handler_marketplace*.go` `handler_settlement.go` `handler_wallet.go` |

---

## 快速开始

```bash
# 前提: Go 1.26+, config.yaml 中配置 API Key
git clone <repo> && cd opc-agent

# CLI 模式
make run

# 桌面端 (macOS)
make desktop && open desktop/build/bin/OPC-Agent.app

# Docker
docker compose up -d
```

---

## Agent 插件

### 现有插件

| 插件 | 技能 | 模型 | HITL | 说明 |
|------|------|------|------|------|
| copywriter | `#文案` `@copywriter` | deepseek-v4-flash | ✅ | 三段式营销文案生成 |
| email_sorter | `#邮件` `@email_sorter` | gemini-2.0-flash | ✅ | 邮件分类+回复建议 |
| xhs_poster | `#小红书` `@xhs_poster` | deepseek-v4-flash | ✅ | 小红书种草笔记 |
| competitive_analysis | `#竞品` `@competitive` | deepseek-v4-flash | ✅ | SWOT 竞品分析 |
| meeting_minutes | `#会议` `@meeting` | deepseek-v4-flash | ✅ | 会议纪要整理 |

### 内容营销插件（已完成 ✅）

| 插件 | 技能 | 模型 | 说明 |
|------|------|------|------|
| publisher | `@publish` | — | 多平台一键发布（抖音/小红书/B站/YouTube 等 12 平台） |
| engage | `@engage` | — | 跨平台自动互动运营（评论回复 / 品牌监测） |
| content_creator | `@create` | — | AI 图片/视频生成（Midjourney / Seedance / Kling） |
| video_script | `@script` | deepseek-v4-flash | AI 视频脚本生成（7 平台 × 7 风格） |
| trend_radar | `@trend` | deepseek-v4-flash | 热点趋势监测 + 选题推荐（6 平台采集） |
| tag_generator | `@tags` | deepseek-v4-flash | AI 标签推荐（热度模型 + 平台差异化策略） |
| monetize | `@monetize` | — | 内容交易市场（接单/结算/钱包/提现/反作弊） |

> 内容营销升级任务完成 238/240（~99%）。详情见 [阶段状态报告](docs/17-阶段状态报告-内容营销升级.md)。

---

## 使用方法

### CLI 命令

```
@<Agent名> <任务>    指定 Agent 处理（如 @copywriter 写文案）
#<技能> <任务>       指定技能处理（如 #文案 推广文案）
<自然语言>            Router 自动识别意图
@a 任务 | @b          A2A 管道 - Agent A 结果传给 B

workflows             列出可用工作流
run <name> <输入>     执行工作流
usage/cost            Token 用量与费用
new                   开始新会话
reload                热重载配置
```

### 工作流编排

```yaml
# workflows/hot_content_pipeline.yaml
steps:
  - id: trend
    agent: trend_radar
    input: "监测抖音和小红书热点，限美食赛道"
  - id: script
    agent: video_script
    input: "根据热榜第一的话题，生成 60 秒抖音带货脚本"
  - id: tags
    agent: tag_generator
    input: "为上一步脚本推荐最佳标签组合"
  - id: publish
    agent: publisher
    input: "将内容发布到抖音和小红书"
```

```bash
> run hot_content_pipeline 今日热点话题
```

---

## 模型 Provider 支持

| Provider | 默认模型 | API 地址 |
|----------|----------|----------|
| OpenAI | gpt-4o | api.openai.com |
| DeepSeek | deepseek-v4-flash | api.deepseek.com |
| Qwen (通义千问) | qwen-turbo | dashscope.aliyuncs.com/compatible-mode |
| Kimi (Moonshot) | moonshot-v1-8k | api.moonshot.cn |
| Claude (Anthropic) | claude-3-5-sonnet | api.anthropic.com |
| Gemini (Google) | gemini-2.0-flash | generativelanguage.googleapis.com |

---

## 部署

```bash
# Docker (全栈)
docker compose up -d

# SaaS 服务 (需要 PostgreSQL)
createdb opc_saas
./build/saas-server -db "postgres://localhost:5432/opc_saas?sslmode=disable"

# Web UI
cd web && npm install && npm run dev

# 桌面端 (macOS)
cd desktop && wails dev
```

---

## 项目结构

```
├── cmd/
│   ├── opc-agent/               # CLI 主程序 (REPL 入口)
│   └── saas-server/             # SaaS HTTP 服务 (多租户后端)
├── internal/
│   ├── agent/                   # Router, Reviewer, Tracker, Cache, Session
│   ├── config/                  # 配置加载 (6 provider 默认端点)
│   ├── harness/                 # Harness 控制 (Role/State/Contract/Guardrail)
│   ├── hitl/                    # HITL 终端确认 (高风险操作拦截)
│   ├── memory/                  # 向量记忆引擎 + pgvector
│   ├── mcp/                     # MCP Server (SSE 传输, Tool/Resource 暴露)
│   ├── platform/                # 多平台 SDK (抖音/小红书/YouTube 等 12 平台)
│   ├── publish/                 # 发布引擎 (队列/重试/定时)
│   ├── engage/                  # 互动引擎 (自动回复/品牌监测/热点)
│   ├── plugins/                 # 内置 Agent 插件注册 (内嵌模式)
│   ├── runtime/                 # Plugin 运行时 + Loader
│   ├── saas/                    # 多租户 SaaS 数据模型与 API
│   └── workflow/                # 工作流编排引擎 (YAML→DAG)
├── plugins/                     # Agent 插件源码 (编译为 .so 动态库)
│   ├── copywriter/              # 文案生成 — 三段式营销文案
│   ├── email_sorter/            # 邮件分类 — 分类+回复建议
│   ├── xhs_poster/              # 小红书内容 — 种草笔记生成
│   ├── competitive_analysis/    # 竞品分析 — SWOT 报告
│   └── meeting_minutes/         # 会议纪要 — 结构化整理
├── desktop/                     # Wails 桌面端 (macOS)
│   ├── app.go                   # Go 后端绑定 (GetDashboardStats, Execute 等)
│   ├── frontend/                # React 前端 (Vite + Tailwind)
│   │   ├── src/
│   │   │   ├── App.jsx          # 5-Tab 导航壳 (Dashboard/Chat/Agents/Usage/Settings)
│   │   │   ├── pages/           # 页面组件
│   │   │   └── components/      # 通用 UI 组件
│   │   └── wails.json
│   └── build/                   # 编译产物 (.app 包)
├── chrome-ext/                  # Chrome 扩展 (Manifest V3, 无需构建)
│   ├── manifest.json
│   ├── background.js            # Service Worker
│   ├── content.js               # 内容注入脚本
│   ├── popup/                   # 弹出窗口
│   ├── options/                 # 选项页面
│   └── icons/                   # 扩展图标
├── web/                         # React Web UI (Vite + Tailwind + Recharts)
│   ├── src/
│   │   ├── components/          # UI 组件
│   │   ├── pages/               # 7 个页面路由
│   │   ├── lib/                 # 工具函数
│   │   └── assets/              # 静态资源
│   └── public/
├── pkg/
│   └── contracts/               # Agent 输入/输出 JSON Schema 契约
├── workflows/                   # 工作流定义 (YAML)
├── data/                        # 记忆持久化存储 (memory.json)
├── scripts/
│   └── create-plugin.sh         # 插件脚手架 (make new-plugin)
├── docs/                        # 设计文档 (18 份)
│   ├── 01 项目需求说明书.md
│   ├── ...
│   ├── 12 内容营销升级需求说明书.md
│   ├── 13 内容创作增强层需求说明书（Phase A）.md
│   ├── 14 商业化变现层需求说明书（Phase C）.md
│   └── 15 内容营销升级用户故事文档.md
├── build/                       # 编译产物
├── Makefile                     # 构建 / 测试 / 运行 / 插件创建
├── Dockerfile                   # 多阶段构建 (Go → Alpine)
├── docker-compose.yml           # 容器编排 (含健康检查)
└── config.yaml                  # 主配置文件 (API Key / Provider / MCP 设置)
```

---

## 开发

```bash
# 创建新插件
make new-plugin NAME=my_agent SUMMARY="说明" TAGS="tag1 tag2"

# 运行测试
make test-all

# 编译全部
make all

# 构建桌面端
make desktop

# 桌面端开发模式 (热重载)
cd desktop && wails dev
```

---

## 版本历史

| 版本 | 日期 | 亮点 |
|------|------|------|
| v0.2.0 | 2026-05 | Wails 桌面端、Chrome 扩展、SaaS 多租户、工作流引擎 |
| v0.1.0 | 2026-04 | CLI MVP、5 个 Agent 插件、6 个模型 Provider、Docker |

---

## 路线图

```
Phase B ───→ Phase A ───→ Phase C
(内容分发层)   (创作增强层)   (商业变现层)

B: 多平台发布 + 自动互动 + MCP 协议 + AI 创作增强
A: 视频脚本生成 + 热点雷达 + AI 标签推荐
C: 内容交易市场 + 结算引擎 + 钱包提现
```

详见 [docs/12 内容营销升级需求说明书](./docs/12%20OPC-Agent%20%E5%86%85%E5%AE%B9%E8%90%A5%E9%94%80%E5%8D%87%E7%BA%A7%E9%9C%80%E6%B1%82%E8%AF%B4%E6%98%8E%E4%B9%A6.md)。
