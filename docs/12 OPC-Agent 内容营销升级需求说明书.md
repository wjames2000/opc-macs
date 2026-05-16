# OPC-Agent 内容营销升级需求说明书

> **版本：** v1.0
> **日期：** 2026-05-13
> **状态：** 草案
> **前置文档：** [01 项目需求说明书](./01%20%E5%A4%9A%E4%BB%BB%E5%8A%A1%E6%99%BA%E8%83%BD%E5%8D%8F%E5%8A%A9%E7%B3%BB%E7%BB%9F%EF%BC%88OPC-Agent%EF%BC%89%E2%80%94%E2%80%94%E9%A1%B9%E7%9B%AE%E9%9C%80%E6%B1%82%E8%AF%B4%E6%98%8E%E4%B9%A6.md)

---

## 1. 概述

### 1.1 项目背景

OPC-Agent 目前已具备多 Agent 编排、多模型 Provider 支持、工作流引擎、SaaS 多租户、Web UI、桌面端和 Chrome 扩展等基础能力。参考开源项目 [AiToEarn](https://github.com/yikart/AiToEarn)（⭐ 10.5k）的设计理念，本次升级将 OPC-Agent 从「AI 任务执行平台」扩展为「AI 内容营销全链路平台」，覆盖内容创作、多平台发布、自动互动运营和内容变现四大领域。

### 1.2 参考项目

| 项目 | Stars | 核心能力 | License |
|------|-------|---------|---------|
| AiToEarn (yikart) | 10.5k | Monetize · Publish · Engage · Create | MIT |

### 1.3 实施路线

```
Phase B ────→ Phase A ────→ Phase C
(6-8 周)      (2-3 周)      (12-16 周)

B: 内容平台核心层（发布 + 互动 + MCP + 创作增强）
A: 内容创作增强层（视频脚本 + 热点 + 标签）
C: 商业化变现层（交易市场 + 结算）
```

---

## 2. Phase B：内容平台核心层（优先级最高）

### 2.1 Publish Agent — 一键多平台发布

#### 2.1.1 业务目标

实现「一次创作，多平台发布」能力，支持 13+ 内容平台的一键同步发布。

#### 2.1.2 支持平台

| 区域 | 平台 | 发布能力 | 接入方式 |
|------|------|---------|---------|
| 国内 | 抖音 (Douyin) | 视频发布 | Open API / Cookie |
| | 小红书 (Xiaohongshu) | 图文/视频发布 | Open API |
| | 快手 (Kuaishou) | 视频发布 | Open API |
| | 哔哩哔哩 (Bilibili) | 视频/动态发布 | Open API |
| | 微信视频号 | 视频发布 | 微信开放平台 |
| | 微信公众号 | 图文发布 | 微信公众平台 API |
| 海外 | TikTok | 视频发布 | TikTok Business API |
| | YouTube | 视频发布 | YouTube Data API v3 |
| | Facebook | 图文/视频 | Graph API |
| | Instagram | 图文/Reels | Instagram Basic Display / Graph API |
| | Threads | 图文发布 | Threads API |
| | Twitter (X) | 图文发布 | Twitter API v2 |
| | Pinterest | 图文/Pin发布 | Pinterest API |
| | LinkedIn | 文章/图文 | LinkedIn API |

#### 2.1.3 功能需求

| ID | 需求 | 描述 | 优先级 |
|----|------|------|--------|
| PUB-01 | 平台账号管理 | 支持多个平台的 OAuth 授权 / API Key 绑定，每个平台可管理多个账号 | P0 |
| PUB-02 | 内容适配器 | 自动将内容格式适配到各平台的格式要求（视频比例、文案长度、标签格式等） | P0 |
| PUB-03 | 发布队列 | 支持批量内容排队发布，支持失败重试和发布状态追踪 | P0 |
| PUB-04 | 发布日历 | 可视化日历界面，可拖拽安排发布时间 | P1 |
| PUB-05 | 定时发布 | 支持指定日期时间自动发布 | P0 |
| PUB-06 | 发布预览 | 发布前在各平台预览效果 | P1 |
| PUB-07 | 内容库 | 集中管理待发布的图文/视频素材 | P1 |
| PUB-08 | 发布日志 | 完整的发布历史记录，含成功/失败/错误详情 | P0 |
| PUB-09 | 多账号矩阵 | 同一个平台可绑定多个账号，选择发布到哪个账号 | P1 |

#### 2.1.4 技术方案

```yaml
新增 Agent:
  name: "publisher"
  summary: "一键多平台内容发布"
  model: 支持多模型（内容适配需要 LLM）
  
新增模块:
  - internal/platform/          # 平台 SDK 封装
      - douyin.go
      - xiaohongshu.go
      - youtube.go
      - tiktok.go
      - ...
  - internal/publish/           # 发布引擎
      - engine.go              # 发布队列、重试、状态
      - adapter.go             # 内容格式适配
      - scheduler.go           # 定时发布
  
数据结构:
  - PlatformAccount:           # 平台账号
      ID, UserID, Platform, AccountName, 
      AccessToken, RefreshToken, ExpiresAt
  - PublishTask:               # 发布任务
      ID, AccountID, Content, Platform,
      Status, ScheduledAt, PublishedAt,
      ErrorMessage, RetryCount
  - ContentItem:               # 内容素材
      ID, Title, Description, MediaURLs,
      Tags, PlatformAdaptations
```

### 2.2 Engage Agent — 自动互动运营

#### 2.2.1 业务目标

自动处理跨平台的用户互动，包括评论回复、品牌监测、热点发现和潜在客户挖掘。

#### 2.2.2 功能需求

| ID | 需求 | 描述 | 优先级 |
|----|------|------|--------|
| ENG-01 | 评论自动回复 | 跨平台自动回复评论，支持关键词触发、白名单/黑名单 | P0 |
| ENG-02 | 智能回复生成 | 使用 LLM 根据评论内容生成个性化回复，保持品牌语气 | P0 |
| ENG-03 | 品牌监测 | 监测各平台提到品牌的帖子/评论，实时告警 | P1 |
| ENG-04 | 热点雷达 | 自动扫描各平台热榜/趋势，生成选题建议报告 | P1 |
| ENG-05 | 评论挖掘 | 分析评论内容，识别高转化意向信号（如「怎么买」「多少钱」） | P1 |
| ENG-06 | 互动数据看板 | 跨平台互动数据汇总（评论数、点赞数、新增关注等） | P0 |
| ENG-07 | 自动点赞/关注 | 按策略自动互动以增加曝光 | P2 |

#### 2.2.3 技术方案

```yaml
新增 Agent:
  name: "engage"
  summary: "跨平台自动互动运营"
  
新增模块:
  - internal/engage/            # 互动引擎
      - reply.go               # 自动回复
      - monitor.go             # 品牌监测
      - trend.go               # 热点雷达
      - mining.go              # 评论挖掘
  
数据流:
  平台 API → 评论采集器 → LLM 分析 → 回复生成器 → 平台 API
```

### 2.3 MCP Server — MCP 协议支持

#### 2.3.1 业务目标

将 OPC-Agent 的所有 Agent 能力通过 MCP（Model Context Protocol）协议暴露，使其可在 Claude Desktop、Cursor、OpenClaw 等支持 MCP 的 AI 客户端中直接调用。

#### 2.3.2 功能需求

| ID | 需求 | 描述 | 优先级 |
|----|------|------|--------|
| MCP-01 | MCP Server 核心 | 基于 MCP 协议规范实现 Service，支持 tools、resources、prompts | P0 |
| MCP-02 | Agent 作为 Tool | 每个 Agent 暴露为一个 MCP Tool，输入为任务描述 | P0 |
| MCP-03 | 工作流作为 Tool | 已定义的工作流也可作为 MCP Tool 调用 | P1 |
| MCP-04 | 资源暴露 | 发布状态、用量数据等作为 MCP Resources 暴露 | P1 |
| MCP-05 | SSE 传输 | 支持 SSE (Server-Sent Events) 传输方式 | P0 |
| MCP-06 | MCP 客户端 SDK | 提供 Node.js / Python 客户端 SDK 示例 | P2 |

#### 2.3.3 技术方案

```yaml
新增模块:
  - internal/mcp/               # MCP 服务
      - server.go              # MCP Server 核心
      - tools.go               # Tool 定义和路由
      - resources.go           # Resource 暴露
      - transport.go           # SSE 传输
  
工具映射示例:
  tool "execute_agent" → App.Execute()
  tool "run_workflow"  → workflow.Engine.Run()
  tool "publish_content" → publisher.Publish()
  resource "usage://stats" → GetUsageStats()
  
配置:
  mcp:
    enabled: true
    port: 8090
    transport: "sse"   # 或 "stdio"
```

### 2.4 Create Agent 增强 — AI 内容创作升级

#### 2.4.1 业务目标

在现有文案生成（copywriter）和小红书（xhs_poster）Agent 基础上，新增 AI 视频和图片生成能力。

#### 2.4.2 功能需求

| ID | 需求 | 描述 | 优先级 |
|----|------|------|--------|
| CR-01 | 图片生成 | 集成 Midjourney / DALL-E / Stable Diffusion，根据描述生成配图 | P0 |
| CR-02 | 视频生成 | 集成 Seedance / 可灵(Kling) / Runway，根据脚本生成短视频 | P0 |
| CR-03 | 内容扩写 | AI 将短内容自动扩写为长文 | P1 |
| CR-04 | 内容缩写 | AI 将长内容压缩为摘要/标题 | P1 |
| CR-05 | 批量生产 | 一个主题批量生成多个变体内容 | P1 |
| CR-06 | 风格迁移 | 将已有内容的风格迁移到其他平台适配格式 | P2 |

#### 2.4.3 技术方案

```yaml
新增 Agent:
  name: "content_creator"
  summary: "AI 图片/视频/文案生成"
  
集成方式:
  图片/视频生成通过外部 API 调用（非 LLM），
  文案增强复用现有模型客户端。
  
新增模块:
  - internal/media/             # 媒体生成
      - image.go               # Image generation
      - video.go               # Video generation
  
外部 API:
  - Midjourney: imagine, vary, upscale
  - Seedance: text-to-video
  - Kling: text-to-video, image-to-video
```

---

## 3. Phase A：内容创作增强层

### 3.1 video_script Agent — 视频脚本生成

#### 3.1.1 业务目标

AI 自动生成适配各平台风格的短视频脚本，包含分镜、话术和视觉效果建议。

#### 3.1.2 功能需求

| ID | 需求 | 描述 | 优先级 |
|----|------|------|--------|
| VS-01 | 脚本生成 | 根据主题和平台生成完整视频脚本 | P0 |
| VS-02 | 分镜输出 | 脚本含分镜描述、时长、画面建议 | P0 |
| VS-03 | 平台适配 | 自动适配抖音/快手/B站/TikTok/YouTube 等平台风格 | P0 |
| VS-04 | 话术风格 | 支持多种话术风格（带货/知识/娱乐/教程） | P1 |
| VS-05 | 批量生成 | 一个主题生成多个脚本变体 | P1 |

### 3.2 trend_radar Agent — 热点雷达

#### 3.2.1 业务目标

自动监测各平台热门话题和趋势，生成选题建议。

#### 3.2.2 功能需求

| ID | 需求 | 描述 | 优先级 |
|----|------|------|--------|
| TR-01 | 热榜采集 | 定时采集各平台热榜数据 | P0 |
| TR-02 | 趋势分析 | 分析话题热度趋势（上升/下降/爆发） | P0 |
| TR-03 | 选题推荐 | 基于账号定位推荐适合创作的热门话题 | P0 |
| TR-04 | 竞品监测 | 监测竞品内容的热门话题分布 | P1 |
| TR-05 | 日报/周报 | 自动生成热点趋势报告 | P1 |

### 3.3 tag_generator Agent — 标签推荐

#### 3.3.1 业务目标

基于内容自动生成优化后的标签/话题，提升内容曝光。

#### 3.3.2 功能需求

| ID | 需求 | 描述 | 优先级 |
|----|------|------|--------|
| TG-01 | 标签生成 | 分析内容自动生成相关标签 | P0 |
| TG-02 | 热搜标签 | 结合热榜数据推荐当前热门标签 | P0 |
| TG-03 | 标签组合优化 | 推荐最佳标签组合策略（大标签+小标签） | P1 |
| TG-04 | 平台差异化 | 不同平台输出不同标签策略 | P1 |

---

## 4. Phase C：商业化变现层

### 4.1 Monetize Agent — 内容交易市场

#### 4.1.1 业务目标

构建内容交易市场，让创作者接商家推广任务，实现 CPS/CPE/CPM 多种结算。

#### 4.1.2 功能需求

| ID | 需求 | 描述 | 优先级 |
|----|------|------|--------|
| MON-01 | 任务发布 | 商家发布推广任务（含预算、要求、结算方式） | P0 |
| MON-02 | 任务广场 | 创作者浏览和接单 | P0 |
| MON-03 | 内容交付 | 创作者提交内容供商家审核 | P0 |
| MON-04 | 结算引擎 | CPS/CPE/CPM 自动结算 | P0 |
| MON-05 | 钱包系统 | 创作者收益提现 | P1 |
| MON-06 | 评价体系 | 商家和创作者互评 | P1 |
| MON-07 | 数据报告 | 推广效果数据报告 | P1 |
| MON-08 | 内容版权 | 内容授权和版权保护 | P2 |

#### 4.1.3 结算模式

| 模式 | 全称 | 适用场景 | 结算依据 |
|------|------|---------|---------|
| CPS | Cost Per Sale | 电商带货 | 实际成交额 |
| CPE | Cost Per Engagement | 品牌推广 | 互动数（点赞/评论/分享） |
| CPM | Cost Per Mille | 曝光推广 | 千次曝光 |

---

## 5. 跨版本公共需求

### 5.1 平台 SDK 层

所有版本共享的平台接入层：

```yaml
internal/platform/
  ├── client.go          # 平台客户端接口
  ├── auth.go            # OAuth / API Key 管理
  ├── douyin/            # 抖音
  ├── xiaohongshu/       # 小红书
  ├── kuaishou/          # 快手
  ├── bilibili/          # B站
  ├── wechat/            # 微信生态
  ├── tiktok/            # TikTok
  ├── youtube/           # YouTube
  ├── facebook/          # Facebook / Instagram
  ├── twitter/           # Twitter (X)
  ├── pinterest/         # Pinterest
  ├── linkedin/          # LinkedIn
  └── threads/           # Threads
```

### 5.2 UI 新增页面

| 页面 | 对应 Phase | 说明 |
|------|-----------|------|
| Publish Dashboard | B | 发布管理、发布日历、账号管理 |
| Engage Dashboard | B | 互动数据、评论管理、品牌监测 |
| MCP Settings | B | MCP 服务配置和状态 |
| Content Studio | A+B | 内容创作编辑器、素材库 |
| Trend Dashboard | A | 热点趋势分析看板 |
| Marketplace | C | 任务广场、任务发布 |
| Wallet | C | 收益管理、提现 |

### 5.3 数据库扩展

```sql
-- Phase B
CREATE TABLE platform_accounts (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  platform VARCHAR(50) NOT NULL,
  account_name VARCHAR(255),
  access_token TEXT,
  refresh_token TEXT,
  expires_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE publish_tasks (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  account_id UUID REFERENCES platform_accounts(id),
  content JSONB NOT NULL,
  platform VARCHAR(50) NOT NULL,
  status VARCHAR(20) DEFAULT 'pending',
  scheduled_at TIMESTAMP,
  published_at TIMESTAMP,
  error_message TEXT,
  retry_count INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Phase C
CREATE TABLE marketplace_tasks (
  id UUID PRIMARY KEY,
  advertiser_id UUID NOT NULL,
  title VARCHAR(200),
  budget DECIMAL(12,2),
  settlement_type VARCHAR(10), -- CPS/CPE/CPM
  requirements TEXT,
  status VARCHAR(20),
  created_at TIMESTAMP DEFAULT NOW()
);
```

### 5.4 API 扩展

```yaml
Phase B API:
  POST   /api/v1/platform/accounts          # 绑定平台账号
  GET    /api/v1/platform/accounts          # 账号列表
  DELETE /api/v1/platform/accounts/:id      # 解绑
  POST   /api/v1/publish                    # 创建发布任务
  GET    /api/v1/publish/tasks              # 发布任务列表
  GET    /api/v1/publish/tasks/:id          # 任务详情
  POST   /api/v1/publish/:id/cancel         # 取消发布
  GET    /api/v1/engage/comments            # 评论列表
  POST   /api/v1/engage/reply               # 自动回复
  GET    /api/v1/engage/stats               # 互动统计
  GET    /api/v1/mcp/config                 # MCP 服务配置

Phase A API:
  POST   /api/v1/create/video-script         # 生成视频脚本
  POST   /api/v1/create/trend-report         # 热点报告
  POST   /api/v1/create/tags                 # 标签推荐

Phase C API:
  POST   /api/v1/marketplace/tasks           # 发布任务
  GET    /api/v1/marketplace/tasks           # 任务列表
  POST   /api/v1/marketplace/tasks/:id/apply # 接单
  POST   /api/v1/marketplace/tasks/:id/deliver  # 交付
  GET    /api/v1/wallet/balance              # 钱包余额
  POST   /api/v1/wallet/withdraw             # 提现
```

---

## 6. 技术栈依赖更新

### 6.1 Go 依赖新增

```yaml
# OAuth 2.0
golang.org/x/oauth2
  
# HTTP Client 增强
github.com/go-resty/resty/v2    # 如需要

# MCP Protocol
github.com/mark3labs/mcp-go     # MCP Go SDK
```

### 6.2 Node.js 依赖新增

```yaml
# Web UI（仅当需要新增前端依赖时）
@tanstack/react-query            # 数据请求
react-big-calendar              # 发布日历
```

---

## 7. 实施优先级矩阵

| 模块 | Phase | 工作量预估 | 依赖 | 可独立交付 |
|------|-------|-----------|------|-----------|
| 平台 SDK (抖音/小红书) | B | 2 周 | 无 | ✅ |
| Publish 核心引擎 | B | 2 周 | 平台 SDK | ❌ |
| MCP Server | B | 1 周 | 无 | ✅ |
| Create 增强 (图片/视频) | B | 2 周 | 无 | ✅ |
| Engage 评论回复 | B | 2 周 | 平台 SDK | ❌ |
| Engage 热点雷达 | B | 1 周 | 无 | ✅ |
| video_script Agent | A | 1 周 | 无 | ✅ |
| trend_radar Agent | A | 1 周 | 平台 SDK(热榜) | ❌ |
| tag_generator Agent | A | 0.5 周 | 无 | ✅ |
| 交易市场 Marketplace | C | 4 周 | 账号系统 | ❌ |
| 结算引擎 | C | 3 周 | 交易市场 | ❌ |
| 钱包系统 | C | 2 周 | 结算引擎 | ❌ |

---

## 8. 风险与应对

| 风险 | 影响 | 概率 | 应对方案 |
|------|------|------|---------|
| 平台 API 限制/变更 | 发布功能可用性 | 高 | 多平台冗余，API 降级策略 |
| OAuth Token 过期 | 账号连接中断 | 中 | 自动刷新 Token 机制 |
| 内容审核规则差异 | 内容被屏蔽 | 中 | 各平台预审核 + 违规内容过滤 |
| MCP 协议版本变化 | 兼容性 | 低 | 版本锁定 + 适配器模式 |
| LLM 幻觉生成不准确标签 | 推荐质量 | 中 | 后过滤 + 人工确认 |
| 结算金额纠纷 | 平台信任 | 中 | 清晰的结算规则 + 争议处理流程 |
