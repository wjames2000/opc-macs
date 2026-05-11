# OPC-Agent

多任务智能协助系统 — 为 OPC（一人公司）创业者打造的 AI Agent 管理平台。

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://golang.org)
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
✅ pgvector 向量存储           ✅ 健康检查 API
✅ 插件脚手架                  ✅ 输出格式化
```

## 快速开始

```bash
# 前提: Go 1.26+, config.yaml 中配置 API Key
git clone <repo> && cd opc-agent

# 编译并运行 (CLI 模式)
make run

# 或使用 Docker
docker compose up -d
```

## Agent 插件

| 插件 | 技能 | 模型 | HITL | 说明 |
|------|------|------|------|------|
| copywriter | `#文案` `@copywriter` | deepseek-v4-flash | ✅ | 三段式营销文案生成 |
| email_sorter | `#邮件` `@email_sorter` | gemini-2.0-flash | ✅ | 邮件分类+回复建议 |
| xhs_poster | `#小红书` `@xhs_poster` | deepseek-v4-flash | ✅ | 小红书种草笔记 |
| competitive_analysis | `#竞品` `@competitive` | deepseek-v4-flash | ✅ | SWOT 竞品分析 |
| meeting_minutes | `#会议` `@meeting` | deepseek-v4-flash | ✅ | 会议纪要整理 |

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
# workflows/customer_complaint.yaml
steps:
  - id: classify
    agent: email_sorter
    input: "${trigger.input}"
  - id: reply
    agent: copywriter
    input: "投诉内容：${classify.output}"
```

```bash
> run customer_complaint 产品有质量问题要求退款
```

## 模型 Provider 支持

| Provider | 默认模型 | API 地址 |
|----------|----------|----------|
| OpenAI | gpt-4o | api.openai.com |
| DeepSeek | deepseek-v4-flash | api.deepseek.com |
| Qwen (通义千问) | qwen-turbo | dashscope.aliyuncs.com/compatible-mode |
| Kimi (Moonshot) | moonshot-v1-8k | api.moonshot.cn |
| Claude (Anthropic) | claude-3-5-sonnet | api.anthropic.com |
| Gemini (Google) | gemini-2.0-flash | generativelanguage.googleapis.com |

## 部署

```bash
# Docker
docker compose up -d

# SaaS 服务 (需要 PostgreSQL)
createdb opc_saas
./build/saas-server -db "postgres://localhost:5432/opc_saas?sslmode=disable"

# Web UI
cd web && npm install && npm run dev
```

## 项目结构

```
├── cmd/
│   ├── opc-agent/           # CLI 主程序 (REPL 入口)
│   └── saas-server/         # SaaS HTTP 服务 (多租户后端)
├── internal/
│   ├── agent/               # Router, Reviewer, Tracker, Cache, Session, ModelClient
│   ├── config/              # 配置加载 (6 provider 默认端点)
│   ├── harness/             # Harness 控制 (Role/State/Contract/Guardrail)
│   ├── hitl/                # HITL 终端确认 (高风险操作拦截)
│   ├── memory/              # 向量记忆引擎 + pgvector
│   ├── plugins/             # 内置 Agent 插件注册 (内嵌模式)
│   ├── runtime/             # Plugin 运行时 + Loader
│   ├── saas/                # 多租户 SaaS 数据模型与 API
│   └── workflow/            # 工作流编排引擎 (YAML→DAG)
├── plugins/                 # Agent 插件源码 (编译为 .so 动态库)
│   ├── copywriter/           # 文案生成 — 三段式营销文案
│   ├── email_sorter/         # 邮件分类 — 分类+回复建议
│   ├── xhs_poster/           # 小红书内容 — 种草笔记生成
│   ├── competitive_analysis/ # 竞品分析 — SWOT 报告
│   └── meeting_minutes/      # 会议纪要 — 结构化整理
├── pkg/
│   └── contracts/            # Agent 输入/输出 JSON Schema 契约
├── workflows/               # 工作流定义 (YAML)
├── web/                     # React Web UI (Vite + Tailwind + Recharts)
│   ├── src/
│   │   ├── components/      # UI 组件
│   │   ├── pages/           # 页面路由
│   │   ├── lib/             # 工具函数
│   │   └── assets/          # 静态资源
│   └── public/
├── data/                    # 记忆持久化存储 (memory.json)
├── scripts/
│   └── create-plugin.sh     # 插件脚手架 (make new-plugin)
├── docs/                    # 设计文档 + 高保真原型
├── build/                   # 编译产物 (二进制 + .so 插件)
├── Makefile                 # 构建 / 测试 / 运行 / 插件创建
├── Dockerfile               # 多阶段构建 (Go → Alpine)
└── docker-compose.yml       # 容器编排 (含健康检查)
```

## 开发

```bash
# 创建新插件
make new-plugin NAME=my_agent SUMMARY="说明" TAGS="tag1 tag2"

# 运行测试
make test-all

# 编译全部
make all
```
