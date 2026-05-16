# OPC-Agent 内容创作增强层需求说明书（Phase A）

> **版本：** v1.0
> **日期：** 2026-05-13
> **状态：** 草案
> **前置条件：** Phase B 核心平台层已上线
> **依赖文档：** [12 内容营销升级需求说明书](./12%20OPC-Agent%20%E5%86%85%E5%AE%B9%E8%90%A5%E9%94%80%E5%8D%87%E7%BA%A7%E9%9C%80%E6%B1%82%E8%AF%B4%E6%98%8E%E4%B9%A6.md)

---

## 1. 概述

Phase A 在 Phase B 已搭建的平台 SDK 和发布引擎基础上，聚焦内容创作的智能化增强。通过 video_script、trend_radar、tag_generator 三个 Agent，构建从选题发现 → 脚本创作 → 标签优化的完整内容生产链路。

### 1.1 定位

```
Phase B: 内容分发层（已有平台 SDK + Publish + Engage）
    ↓
Phase A: 内容创作增强层（新增: video_script + trend_radar + tag_generator）
    ↓
Phase C: 内容变现层（待实现）
```

### 1.2 三个 Agent 的协作关系

```
trend_radar ──→ video_script ──→ tag_generator
   │                │                │
   ▼                ▼                ▼
  选题灵感        脚本输出         标签组合
  （驱动创作）     （核心产出）     （曝光优化）
```

---

## 2. video_script Agent — AI 视频脚本生成

### 2.1 Agent 定义

```yaml
name: "video_script"
summary: "AI 自动生成适配各平台风格的短视频脚本"
agent_type: "plugin"        # 复用现有 AgentPlugin 接口
system_prompt: "你是一个专业的短视频脚本创作专家..."
model_required: true        # 需要 LLM 推理
```

### 2.2 输入参数

```yaml
inputs:
  - topic: string           # 创作主题 / 关键词（必填）
  - platform: enum          # 目标平台（可选，默认为通用）
    values: [douyin, kuaishou, bilibili, tiktok, youtube, xiaohongshu, wechat]
  - style: enum             # 话术风格（可选，默认为自动匹配）
    values: [带货, 知识科普, 娱乐搞笑, 教程分享, 产品测评, Vlog, 剧情]
  - duration: string        # 目标时长（可选，如 "60s", "3min"）
  - reference: string       # 参考内容 / 竞品分析（可选）
  - brand_info: string      # 品牌/产品信息（可选，带货类必填）
  - count: int              # 生成数量（可选，默认 1，最大 5）
```

### 2.3 输出格式

```yaml
output:
  scripts:
    - title: string           # 脚本标题
      platform: string        # 适配平台
      duration: string        # 预估时长
      hook: string            # 黄金前 3 秒 / 开场钩子
      scenes:
        - scene_no: int       # 分镜序号
          duration: string    # 该分镜时长
          visual: string      # 画面描述
          audio: string       # 配音内容 / 旁白
          text_overlay: string # 字幕文案
          camera: string      # 运镜建议（推/拉/摇/移/跟）
          b_roll: string      # 辅助画面建议
      cta: string             # 结尾引导（关注/点赞/购买/评论）
      tags: string[]          # 推荐话题标签
  metadata:
    total_duration: string    # 总时长校验
    scene_count: int          # 分镜数
    word_count: int           # 总文案字数
```

### 2.4 平台适配规则

| 平台 | 推荐时长 | 比例 | 话术风格 | 标签数 |
|------|---------|------|---------|--------|
| 抖音 | 15-60s | 9:16 | 前3秒钩子+快节奏 | 3-5 |
| 快手 | 30-90s | 9:16 | 接地气、老铁文化 | 3-5 |
| B站 | 3-15min | 16:9 | 深度、趣聊、梗文化 | 5-10 |
| TikTok | 15-60s | 9:16 | 英文/多语言 | 3-5 |
| YouTube | 8-20min | 16:9 | 信息密度+故事性 | 5-15 |
| 小红书 | 30-120s | 3:4 | 种草、生活化 | 3-5 |

### 2.5 UI 设计

```yaml
页面: Content Studio → video_script 面板

布局:
  左侧配置区:
    - 主题输入框（必填）
    - 平台选择器（多选或全平台）
    - 风格选择（下拉菜单）
    - 时长滑块（15s - 20min）
    - 品牌信息输入（富文本）
    - 参考链接输入
    - 「生成脚本」按钮
  
  右侧结果区:
    - 脚本列表（Tab 切换多个变体）
    - 分镜时间线（可视化，可拖拽调整顺序）
    - 预览模式（文案预览 / 分镜视图）
    - 导出按钮（复制文案 / 导出为模板）
    - 「一键创建发布任务」→ 跳转 Publish Agent
```

### 2.6 集成到工作流

```yaml
# 示例：抖音带货脚本生成工作流
workflow:
  name: "douyin_selling_script"
  steps:
    - id: step1
      agent: trend_radar
      input: "今日抖音热榜 TOP10"
    - id: step2
      agent: video_script
      input: "根据热榜第1名话题生成60秒带货脚本"
    - id: step3
      agent: tag_generator
      input: "为上一步脚本推荐最佳标签组合"
    - id: step4
      agent: publisher
      input: "将脚本关联素材发布到抖音账号"
```

---

## 3. trend_radar Agent — 热点雷达

### 3.1 Agent 定义

```yaml
name: "trend_radar"
summary: "自动监测各平台热门话题和趋势，生成选题建议"
agent_type: "plugin"
system_prompt: "你是一个社交媒体趋势分析专家..."
model_required: true
```

### 3.2 功能模块

#### 3.2.1 热榜采集

| 数据源 | 采集方式 | 频率 | 数据量 |
|--------|---------|------|--------|
| 抖音热榜 | Web Scraping / API | 每 30min | TOP 50 |
| 小红书热搜 | Web Scraping | 每 30min | TOP 30 |
| B站热门 | B站 API | 每 30min | TOP 100 |
| 微博热搜 | Web Scraping | 每 30min | TOP 50 |
| 快手热榜 | Web Scraping | 每 30min | TOP 30 |
| TikTok Trends | Web Scraping | 每 60min | TOP 20 |
| YouTube Trends | YouTube API | 每 60min | TOP 20 |
| Twitter Trends | Twitter API | 每 60min | TOP 10 (by region) |

#### 3.2.2 输入参数

```yaml
inputs:
  - platforms: enum[]      # 监测平台列表（必填）
  - niche: string          # 垂直领域/赛道（可选，用于过滤相关性）
  - days_back: int         # 回顾天数（可选，默认 1）
  - generate_report: bool  # 是否生成报告（可选，默认 false）
  - competitor: string     # 竞品账号名（可选，竞品监测模式）
```

#### 3.2.3 输出格式

```yaml
output:
  hot_topics:
    - rank: int
      title: string             # 话题标题
      platform: string          # 来源平台
      heat_score: int           # 热度指数（0-100）
      trend: enum               # 趋势方向
        values: [rising, peaking, declining, stable]
      related_tags: string[]    # 关联标签
      summary: string           # LLM 生成的话题摘要
      
  recommendations:
    - topic: string             # 推荐选题
      reason: string            # 推荐理由（热度+相关性+竞争度）
      suggested_platform: string
      confidence: float         # 置信度（0-1）
      
  # 当 generate_report=true 时额外输出
  report:
    period: string              # 报告周期
    overview: string            # 趋势总览
    top_5: string[]             # 最值得关注的话题
    niche_insights: string      # 垂直领域洞察
    competitor_analysis: string # 竞品分析（限竞品监测模式）
```

### 3.3 UI 设计

```yaml
页面: Trend Dashboard

布局:
  顶部筛选栏:
    - 平台选择器（多选）→ 影响下面的数据面板
    - 时间范围选择（24h / 3d / 7d）
    - 垂直领域输入
  
  热榜面板:
    - 话题列表（排名 + 话题 + 热度趋势图 + 平台标识）
    - 点击话题展开详情（LLM 摘要 + 关联账号 + 相关标签）
    - 一键「以此创作」→ 跳转 video_script
  
  趋势图表:
    - 热度趋势折线图（随时间变化）
    - 平台对比柱状图
    - 热门词云
  
  推荐选题:
    - 卡片式展示（话题 + 平台 + 置信度 + 推荐理由）
    - 「采纳选题」→ 创建创作任务
```

---

## 4. tag_generator Agent — AI 标签推荐

### 4.1 Agent 定义

```yaml
name: "tag_generator"
summary: "基于内容自动生成优化后的标签/话题标签，提升内容曝光"
agent_type: "plugin"
system_prompt: "你是一个标签优化专家，精通各平台搜索和推荐算法..."
model_required: true
```

### 4.2 输入参数

```yaml
inputs:
  - content: string         # 内容标题/描述/全文（必填）
  - platform: enum          # 目标平台（必填）
    values: [douyin, xiaohongshu, kuaishou, bilibili, 
             tiktok, youtube, twitter, instagram]
  - category: string        # 内容分类（可选）
  - existing_tags: string[] # 已有标签（可选，基础上优化）
  - count: int              # 推荐数量（可选，默认 10，最大 20）
  - include_hot: bool       # 是否结合热搜（可选，默认 true）
  - competitor_tags: string[] # 竞品标签（可选，参考优化）
```

### 4.3 输出格式

```yaml
output:
  primary_tags:              # 核心标签（高流量、高相关）
    - tag: string            # 标签名
      heat_level: enum       # 热度等级
        values: [🔥火爆, 📈上升, 💧长尾]
      estimated_reach: string # 预估曝光（参考值）
      reason: string         # 推荐理由
      
  secondary_tags:            # 辅助标签（精准流量）
    - tag: string
      heat_level: enum
      estimated_reach: string
      reason: string
      
  strategy:                  # 标签策略说明
    combination: string      # 推荐组合方式
    tips: string[]           # 优化建议
    avoid: string[]          # 不推荐的标签
```

### 4.4 标签策略模型

```
策略层级          示例（抖音带货）
─────────────────────────────────────
流量标签       →  #好物推荐 #好物分享
品类标签       →  #零食测评 #零食推荐
场景标签       →  #办公室零食 #追剧零食
人群标签       →  #吃货 #宝妈必看
品牌标签       →  #三只松鼠 #良品铺子
热点标签       →  #618必买 #双11清单
```

### 4.5 UI 设计

```yaml
页面: Content Studio → tag_generator 面板（或嵌入发布步骤）

布局:
  输入区:
    - 内容简介输入框
    - 平台选择器
    - 分类选择
    - 现有标签输入（可选）
    - 「生成标签」按钮
  
  结果区:
    - 标签云图（按热度大小排列）
    - 核心标签列表（带热度标识和理由）
    - 辅助标签列表
    - 策略建议卡片
    - 「一键复制标签」按钮
    - 「添加到发布任务」按钮
```

---

## 5. 三个 Agent 的集成与协作

### 5.1 在 Create Agent 中的角色

```yaml
# 原 Create Agent（Phase B 增强后）
create_agent:
  capabilities:
    - text_generation: copywriter (已有的文案生成)
    - image_generation: Phase B 增强
    - video_generation: Phase B 增强
    
# Phase A 新增子能力（作为 Create 的扩展）
create_agent_v2:
  capabilities:
    - text_generation: copywriter
    - image_generation: Phase B
    - video_generation: Phase B
    - video_script: video_script Agent  ← NEW
    - trend_analysis: trend_radar Agent  ← NEW
    - tag_optimization: tag_generator Agent  ← NEW
```

### 5.2 工作流集成示例

```yaml
name: "全自动内容创作流水线"
description: "热点分析 → 脚本生成 → 标签优化 → 自动发布"

steps:
  - id: step1
    agent: trend_radar
    input: "监测抖音和小红书热点，限美食赛道"
    
  - id: step2
    agent: video_script
    input: "根据热榜第一的话题，生成60秒抖音脚本，带货风格，品牌：XX零食"
    
  - id: step3
    agent: tag_generator
    input: "为step2脚本推荐抖音最佳标签组合"
    
  - id: step4
    agent: content_creator
    input: "根据step2脚本生成短视频（调用Kling API）"
    
  - parallel:
    - id: step5a
      agent: publisher
      input: "发布抖音"
    - id: step5b
      agent: publisher
      input: "发布小红书"
```

### 5.3 API 接口

```yaml
POST /api/v1/create/video-script
  Request:  { topic, platform?, style?, duration?, reference?, brand_info?, count? }
  Response: { scripts: [...], metadata: {...} }

POST /api/v1/create/trend-report
  Request:  { platforms, niche?, days_back?, generate_report?, competitor? }
  Response: { hot_topics: [...], recommendations: [...], report?: {...} }

POST /api/v1/create/tags
  Request:  { content, platform, category?, existing_tags?, count?, include_hot? }
  Response: { primary_tags: [...], secondary_tags: [...], strategy: {...} }
```

---

## 6. 与 Phase B 的依赖关系

| Phase A 模块 | 依赖 Phase B | 说明 |
|-------------|-------------|------|
| video_script | 无硬依赖 | 独立 Agent，仅需要 LLM |
| video_script 的「一键发布」| Publish Agent | UI 跳转功能，非核心依赖 |
| trend_radar (热榜采集) | 平台 SDK | 需要抖音/小红书等平台 SDK（Phase B 实现）|
| trend_radar (Web Scraping 模式) | 无硬依赖 | 可先用无头浏览器方案 |
| tag_generator | 无硬依赖 | 独立 Agent，仅需要 LLM |
| tag_generator (热搜标签) | trend_radar | 需要热榜数据作为输入 |

### 6.1 无依赖组件的独立交付顺序

```
第一周：video_script Agent（独立可交付）
  ├── 插件注册（复用 AgentPlugin 接口）
  ├── 核心脚本生成逻辑
  ├── 平台风格适配
  └── UI 面板

第二周：tag_generator Agent（独立可交付）
  ├── 插件注册
  ├── 标签生成逻辑
  ├── 热搜标签集成（依赖 Phase B trend_radar 则有，否则纯 LLM）
  └── UI 面板

第三周：trend_radar Agent（部分依赖 Phase B 平台 SDK）
  ├── Web Scraping 数据采集（独立方案）
  ├── LLM 分析引擎
  ├── 选题推荐逻辑
  ├── 报告生成
  └── UI 面板（Trend Dashboard）
```

---

## 7. 工作量预估

| 模块 | 后端（天）| 前端（天）| 测试（天）| 合计 |
|------|---------|---------|---------|------|
| video_script Agent | 3 | 2 | 1 | 6 |
| tag_generator Agent | 2 | 1.5 | 0.5 | 4 |
| trend_radar 数据采集 | 3 | - | 1 | 4 |
| trend_radar 分析引擎 | 2 | - | 1 | 3 |
| trend_radar UI (Trend Dashboard) | - | 3 | 1 | 4 |
| 集成测试 + 文档 | 1 | 1 | 1 | 3 |
| **总计** | **11** | **7.5** | **5.5** | **24** |

---

## 8. 验收标准

### 8.1 video_script

- [ ] 输入主题后，5 秒内生成完整视频脚本
- [ ] 生成脚本含完整分镜、话术、画面建议
- [ ] 支持 7+ 个平台的风格适配
- [ ] 支持 7+ 种话术风格
- [ ] 一键导出 / 复制脚本文案
- [ ] LLM 生成的脚本长度符合目标平台推荐时长

### 8.2 trend_radar

- [ ] 每 30 分钟自动采集 6+ 平台热榜数据
- [ ] 正确识别话题热度趋势方向（上升/峰值/下降/稳定）
- [ ] 选题推荐置信度 > 70% 接受率（用户主观评价）
- [ ] 热点日报/周报格式完整可读
- [ ] 支持垂直领域过滤

### 8.3 tag_generator

- [ ] 生成标签与内容相关性 > 85%
- [ ] 准确区分核心标签和辅助标签
- [ ] 标签热度标识与实际平台趋势一致
- [ ] 支持 8+ 个平台的标签策略差异化
- [ ] 标签组合策略建议具有实际可操作性
