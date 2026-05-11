# OPC-Agent SaaS 多租户架构方案

> **版本**：v1.0  
> **日期**：2026-05-11  
> **项目代号**：OPC-Agent

---

## 1. 方案概述

将 OPC-Agent 从单用户 CLI 工具升级为多租户 SaaS 平台，让多个团队/个人共用一套部署，各自独立使用 Agent 能力，互不干扰。

### 核心目标

| 目标 | 说明 |
|------|------|
| **多租户隔离** | 租户间数据完全隔离，互不可见 |
| **按需计费** | 按 Token 消耗 + 用户数 + 插件数计费 |
| **自助注册** | 用户可自助注册、配置、使用 |
| **弹性扩缩** | 根据租户数量自动扩缩容 |

---

## 2. 总体架构

```
┌─────────────────────────────────────────────────────────┐
│                   客户端层                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐               │
│  │ Web UI   │  │ API SDK  │  │ CLI      │               │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘               │
├───────┴─────────────┴─────────────┴────────────────────┤
│                    API 网关层                            │
│  ┌──────────────────────────────────────────────────┐  │
│  │  API Gateway / Load Balancer                     │  │
│  │  ┌──────────┐ ┌──────────┐ ┌───────────────┐     │  │
│  │  │ 认证鉴权  │ │ 限流     │ │ 租户路由      │     │  │
│  │  └──────────┘ └──────────┘ └───────────────┘     │  │
│  └──────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────┤
│                    业务服务层                            │
│  ┌──────────────────────────────────────────────────┐  │
│  │  User Service     Tenant Service    Billing Svc  │  │
│  │  (用户管理)        (租户配置)        (计费)       │  │
│  ├──────────────────────────────────────────────────┤  │
│  │  Agent Orchestrator   Workflow Engine            │  │
│  │  (Agent 编排/租户隔离)  (工作流执行)              │  │
│  ├──────────────────────────────────────────────────┤  │
│  │  Model Router        Token Tracker               │  │
│  │  (模型路由/租户Key)   (用量追踪/计费)             │  │
│  └──────────────────────────────────────────────────┘  │
├────────────────────────────────────────────────────────┤
│                    数据层                                │
│  ┌──────────┐  ┌──────────┐  ┌────────────────────┐   │
│  │ Postgres │  │  Redis   │  | 对象存储 (S3)       │   │
│  │ (业务数据)│  │ (会话/缓存)│  │ (用户文件/Agent产出)│   │
│  │ +pgvector│  │          │  │                    │   │
│  └──────────┘  └──────────┘  └────────────────────┘   │
└────────────────────────────────────────────────────────┘
```

---

## 3. 多租户数据模型

### 3.1 核心表结构 (PostgreSQL)

```sql
-- ===== 租户 =====
CREATE TABLE tenants (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,           -- 租户名称
    slug        TEXT UNIQUE NOT NULL,    -- 子域名标识
    plan        TEXT NOT NULL DEFAULT 'free',  -- free | pro | enterprise
    status      TEXT NOT NULL DEFAULT 'active', -- active | suspended | cancelled
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settings    JSONB DEFAULT '{}',      -- 租户级配置
    api_key     TEXT UNIQUE              -- 租户 API Key
);

-- ===== 用户（租户内） =====
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    email       TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'member', -- owner | admin | member
    status      TEXT NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

-- ===== 租户模型配置 =====
CREATE TABLE tenant_models (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    provider    TEXT NOT NULL,           -- openai | deepseek | etc
    model_name  TEXT NOT NULL,
    api_key     TEXT NOT NULL,           -- 加密存储
    api_base_url TEXT,
    is_default  BOOLEAN DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ===== 租户 Agent 配置 =====
CREATE TABLE tenant_agents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    agent_name  TEXT NOT NULL,           -- copywriter | email_sorter | custom
    model_id    UUID REFERENCES tenant_models(id),
    enabled     BOOLEAN DEFAULT true,
    config      JSONB DEFAULT '{}',     -- Temperature, MaxTokens, Prompt
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ===== Token 用量记录 =====
CREATE TABLE token_usage (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    user_id     UUID REFERENCES users(id),
    agent_name  TEXT NOT NULL,
    model_name  TEXT NOT NULL,
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    cost        DECIMAL(12,6) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- ===== 计费记录 =====
CREATE TABLE billing_records (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    period_start DATE NOT NULL,
    period_end   DATE NOT NULL,
    total_tokens BIGINT NOT NULL DEFAULT 0,
    total_cost   DECIMAL(12,4) NOT NULL DEFAULT 0,
    plan_fee     DECIMAL(12,4) NOT NULL DEFAULT 0,
    status       TEXT NOT NULL DEFAULT 'pending', -- pending | paid | overdue
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 3.2 租户隔离策略

```go
// 所有查询强制携带 tenant_id
type TenantContext struct {
    TenantID string
    UserID   string
    Role     string
}

// 中间件自动注入租户上下文
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID := extractTenantID(r) // 从 JWT/Subdomain/API Key 提取
        ctx := context.WithValue(r.Context(), "tenant", &TenantContext{
            TenantID: tenantID,
        })
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Repository 层自动过滤
type AgentRepository struct {
    db *sql.DB
}

func (r *AgentRepository) List(ctx context.Context) ([]Agent, error) {
    tc := ctx.Value("tenant").(*TenantContext)
    rows, _ := r.db.QueryContext(ctx,
        "SELECT * FROM tenant_agents WHERE tenant_id = $1", tc.TenantID)
    // ...
}
```

---

## 4. 租户识别方案

| 方式 | 适用场景 | 实现 |
|------|----------|------|
| **子域名** | Web UI | `tenantA.opc-agent.com` → Header → Tenant ID |
| **API Key** | 程序化调用 | `Authorization: Bearer tk_xxxxx` → 查 tenants.api_key |
| **JWT Token** | 用户登录 | JWT payload 含 `tenant_id` + `user_id` |
| **Header** | 企业集成 | 自定义 Header `X-Tenant-ID` |

### 路由策略

```
                            ┌─ tenant-a.opc-agent.com ──► Tenant A
                           ├─ tenant-b.opc-agent.com ──► Tenant B
  API Gateway ──► 域名解析 ──┼─ ...
                            │
                            └─ api.opc-agent.com ──► API Key → Tenant
```

---

## 5. 计费模型

### 5.1 套餐方案

| 套餐 | 价格 | Token 配额 | Agent 数 | 用户数 | 工作流 |
|------|------|-----------|----------|--------|--------|
| **Free** | ¥0 | 10 万/月 | 3 | 1 | 2 |
| **Pro** | ¥99/月 | 100 万/月 | 10 | 5 | 20 |
| **Enterprise** | ¥499/月 | 1000 万/月 | 不限 | 不限 | 不限 |

### 5.2 超额计费

```yaml
# 超出套餐配额后按量计费
overage:
  token_price: ¥0.008/千 tokens    # DeepSeek 等性价比模型
  premium_token_price: ¥0.03/千    # GPT-4/Claude 等高端模型
  additional_user: ¥10/用户/月
  additional_agent: ¥5/Agent/月
```

### 5.3 用量统计与账单

```
┌────────────────────────────────────────────────────┐
│ 📊 本月用量                                       │
├────────────────────────────────────────────────────┤
│  套餐: Pro ¥99/月                                   │
│  Token: 845,200 / 1,000,000 (84.5%)                │
│  ███████████████████████░░░░░░░░                    │
│  超额: 0                                           │
│  预估费用: ¥99.00                                   │
├────────────────────────────────────────────────────┤
│  按 Agent 统计:                                    │
│  copywriter:       345,200 tokens    ¥2.76         │
│  email_sorter:     210,000 tokens    ¥1.68         │
│  xhs_poster:       180,000 tokens    ¥1.44         │
│  competitive:      110,000 tokens    ¥0.88         │
├────────────────────────────────────────────────────┤
│  [📥 下载账单]  [⚙️ 升级套餐]                      │
└────────────────────────────────────────────────────┘
```

---

## 6. API 设计

### 6.1 RESTful API

```yaml
openapi: 3.0.0
info:
  title: OPC-Agent SaaS API
  version: v1

paths:
  # ===== 认证 =====
  /api/v1/auth/register:
    post: 用户注册（自动创建租户）
  /api/v1/auth/login:
    post: 用户登录（返回 JWT）

  # ===== Agent 执行 =====
  /api/v1/agents:
    get: 获取可用 Agent 列表
  /api/v1/agents/{name}/execute:
    post: |
      执行 Agent 任务
      Body: { "input": "..." }
      Response: { "result": {...}, "tokens": {...} }

  # ===== 工作流 =====
  /api/v1/workflows:
    get: 获取工作流列表
    post: 创建工作流
  /api/v1/workflows/{id}/run:
    post: 执行工作流
  /api/v1/workflows/{id}/status:
    get: 查询执行状态

  # ===== 用量 =====
  /api/v1/usage:
    get: 获取用量统计（按时间/Agent/模型）

  # ===== 管理 =====
  /api/v1/tenant/models:
    get: 获取租户模型列表
    post: 添加模型配置
  /api/v1/tenant/agents:
    get: 获取 Agent 配置列表
    put: 更新 Agent 配置
```

### 6.2 SDK 示例

```python
# Python SDK
from opc_agent import OPCAgent

client = OPCAgent(api_key="tk_xxxxxxxx")

# 执行 Agent
result = client.agents.execute("copywriter", "写一个产品文案")
print(result.data)

# 执行工作流
wf = client.workflows.run("customer_complaint", "我要退款")
print(wf.status)
```

```javascript
// JavaScript SDK
import { OPCClient } from 'opc-agent-sdk';

const client = new OPCClient({ apiKey: 'tk_xxxx' });
const result = await client.execute('copywriter', '写文案');
```

---

## 7. 部署架构

```
┌─────────────────────────────────────────────────────┐
│  Kubernetes Cluster                                 │
│                                                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │  Web UI     │  │  API Server │  │  Admin Panel│  │
│  │  (React)    │  │  (Go)       │  │  (React)    │  │
│  └─────────────┘  └──────┬──────┘  └─────────────┘  │
│                          │                           │
│  ┌───────────────────────┴──────────────────────┐   │
│  │  Agent Worker Pool (水平扩展)                  │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐     │   │
│  │  │ Worker 1 │ │ Worker 2 │ │ Worker N │     │   │
│  │  └──────────┘ └──────────┘ └──────────┘     │   │
│  └──────────────────────────────────────────────┘   │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
│  │ Postgres │  │  Redis   │  │ 对象存储 (S3)     │  │
│  │ +pgvector│  │ (会话)   │  │ (文件/日志)       │  │
│  └──────────┘  └──────────┘  └──────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### 水平扩展策略

| 组件 | 扩展方式 | 说明 |
|------|----------|------|
| API Server | 无状态，水平扩展 | 基于 K8s HPA，CPU > 70% 自动扩容 |
| Agent Worker | 任务队列驱动 | 每个租户的任务分发到 Worker 池 |
| Postgres | 读写分离 + 分片 | token_usage 按时间分区，tenant 级分片 |
| Redis | Cluster 模式 | 会话缓存 + 任务队列 + 限流计数器 |

---

## 8. 安全设计

| 维度 | 措施 |
|------|------|
| **数据隔离** | 所有 SQL 查询强制 `WHERE tenant_id = $1` |
| **API Key 管理** | Key 加密存储（AES-256），支持轮换 |
| **限流** | 每租户每秒 N 次请求，超额 429 |
| **审计日志** | 所有 API 调用记录，保留 90 天 |
| **数据加密** | 传输 TLS 1.3，静态 AES-256 |
| **Key 加密存储** | 使用 Vault/KMS 加密租户的模型 API Key |

---

## 9. 实施路线

| 阶段 | 内容 | 工时 |
|------|------|------|
| **Phase 1** | 多租户数据模型 + 租户识别中间件 | 2d |
| **Phase 2** | 用户认证 + 租户管理 API | 2d |
| **Phase 3** | Agent/Workflow 多租户隔离改造 | 2d |
| **Phase 4** | Token 计费追踪 + 配额管理 | 1.5d |
| **Phase 5** | Web UI 多租户适配 + 管理员面板 | 2d |
| **Phase 6** | K8s 部署 + 水平扩展 + 监控 | 2d |

---

*文档结束*
