# OPC-Agent 商业化变现层需求说明书（Phase C）

> **版本：** v1.0
> **日期：** 2026-05-13
> **状态：** 草案
> **前置条件：** Phase B 内容平台核心层 + Phase A 内容创作增强层已上线
> **依赖文档：** [12 内容营销升级需求说明书](./12%20OPC-Agent%20%E5%86%85%E5%AE%B9%E8%90%A5%E9%94%80%E5%8D%87%E7%BA%A7%E9%9C%80%E6%B1%82%E8%AF%B4%E6%98%8E%E4%B9%A6.md)

---

## 1. 概述

Phase C 将 OPC-Agent 从「内容生产工具」升级为「内容交易平台」。构建连接商家与创作者的交易市场，支持 CPS（按成交付费）、CPE（按互动付费）、CPM（按曝光付费）三种结算模式，同时提供完整的钱包、提现和数据分析体系。

### 1.1 核心角色

| 角色 | 定义 | 核心诉求 |
|------|------|---------|
| 商家（Advertiser）| 有推广需求的品牌/商家 | 找到优质创作者，按效果付费 |
| 创作者（Creator）| 已接入平台账号的内容创作者 | 接赚钱任务，高效变现 |
| 平台（Platform）| OPC-Agent 运营方 | 抽佣，保障交易可信 |

### 1.2 业务闭环

```
商家发布任务 ──→ 创作者接单 ──→ 内容制作 ──→ 发布推广
                                                    │
               收益提现 ←── 结算引擎 ←── 效果归因 ←──┘
```

---

## 2. 交易市场核心功能

### 2.1 任务管理

#### 2.1.1 商家视角

| ID | 需求 | 描述 | 优先级 |
|----|------|------|--------|
| MKT-01 | 任务发布 | 填写任务信息：标题、预算、要求、结算方式、截止日期 | P0 |
| MKT-02 | 模板任务 | 预设常见任务模板（带货推广/品牌曝光/新品预热） | P1 |
| MKT-03 | 创作者筛选 | 按粉丝量、垂直领域、平均互动率筛选创作者 | P0 |
| MKT-04 | 定向邀请 | 向指定创作者发送任务邀请 | P1 |
| MKT-05 | 任务审核 | 查看创作者提交的内容，批准/驳回/要求修改 | P0 |
| MKT-06 | 数据报告 | 查看推广效果数据（曝光/互动/转化/ROI） | P0 |
| MKT-07 | 任务管理 | 查看发布中/进行中/已完成/已取消的任务列表 | P0 |

#### 2.1.2 任务发布表单

```yaml
字段:
  title: string                # 任务标题（必填）
  description: string          # 任务描述（必填）
  platform: string[]           # 目标平台（必填，可多选）
  settlement_type: enum        # 结算方式（必填）
    values: [CPS, CPE, CPM]
  budget:
    type: number               # 总预算（必填）
    unit: string               # 货币单位（CNY/USD）
  requirements:
    min_followers: number      # 最低粉丝数
    min_engagement: number     # 最低互动率(%)
    niche: string[]            # 垂直领域要求
    content_type: string[]     # 内容类型：video/image/text
    guidelines: string         # 创作要求说明
  deadline: datetime           # 截止日期
  max_creators: int            # 最大接单人数
```

### 2.2 任务广场（创作者视角）

| ID | 需求 | 描述 | 优先级 |
|----|------|------|--------|
| MKT-08 | 任务浏览 | 按平台/结算方式/预算范围/领域筛选任务 | P0 |
| MKT-09 | 智能推荐 | 根据创作者的账号定位和过往内容推荐匹配任务 | P1 |
| MKT-10 | 一键接单 | 满足条件即可直接接单 | P0 |
| MKT-11 | 内容交付 | 提交已发布内容的链接和数据截图 | P0 |
| MKT-12 | 收益查看 | 查看已结算/待结算收益 | P0 |

#### 2.2.1 任务卡片

```yaml
TaskCard:
  封面: 品牌 Logo / 任务类型图标
  标题: 任务标题
  商家: 品牌名 + 认证标识
  标签: [平台] [结算方式] [预算范围]
  数据:
    - 预算: ¥5,000
    - 已接单: 3/10
    - 截止: 2026-06-15
  要求摘要:
    - 粉丝 ≥ 10,000
    - 抖音 / 小红书
    - 互动率 ≥ 3%
  操作: 【查看详情】| 【立即接单】
```

### 2.3 接单与交付流程

```
接单阶段
  创作者点击「接单」→ 系统检查是否符合条件
    → 符合：锁定名额 → 进入创作阶段
    → 不符合：显示不满足的条件

创作阶段（7-14 天）
  创作者制作内容 → 发布到目标平台
    → 在系统内提交发布链接和效果截图

审核阶段（1-3 天）
  商家审核交付内容
    ├── 批准 → 进入结算
    ├── 驳回 → 创作者修改后重新提交（最多 3 次）
    └── 争议 → 平台介入仲裁

结算阶段
  系统根据结算方式自动计算金额
    → 平台扣除佣金 → 金额记入创作者钱包
```

---

## 3. 结算引擎

### 3.1 三种结算模型

#### 3.1.1 CPS（Cost Per Sale）— 按成交付费

```
结算金额 = 实际成交额 × CPS 费率

示例：
  商品售价：¥100
  商家设定 CPS 费率：20%
  创作者促成 50 单
  结算金额 = 100 × 20% × 50 = ¥1,000
```

**归因方式：**
- 抖音：电商联盟 PID / 小程序下单
- 小红书：商品链接 / 笔记挂车
- B站：商品橱窗 / 邀约广告

#### 3.1.2 CPE（Cost Per Engagement）— 按互动付费

```
结算金额 = 有效互动数 × 单次互动价格

示例：
  单次互动出价：¥0.5
  创作者内容获得：
    点赞：2,000
    评论：500
    分享：200
  有效互动 = 2,000 × 0.5 + 500 × 1 + 200 × 1.5（加权）
  结算金额 = 有效互动数 × ¥0.5 = ...
```

**互动加权：**
| 互动类型 | 权重 | 说明 |
|---------|------|------|
| 点赞 | 1× | 最轻量互动 |
| 收藏 | 1.5× | 兴趣信号 |
| 评论 | 2× | 深度互动 |
| 分享 | 3× | 传播行为 |

#### 3.1.3 CPM（Cost Per Mille）— 按曝光付费

```
结算金额 = (总曝光量 / 1,000) × CPM 单价

示例：
  CPM 出价：¥50（每千次曝光 ¥50）
  内容获得：100,000 次曝光
  结算金额 = (100,000 / 1,000) × ¥50 = ¥5,000
```

#### 3.1.4 结算方式选择策略

```yaml
推荐场景:
  CPS: 带货类任务（有明确转化目标）
  CPE: 品牌互动 / 话题挑战
  CPM: 品牌曝光 / 新品预热
  混合: 保底 CPM + 激励 CPS（高预算任务）
```

### 3.2 结算流程

```yaml
步骤:
  1. 商家创建任务时设定结算方式和预算
  2. 创作者完成交付并获批准
  3. 系统根据归因数据计算最终金额
  4. 平台按 x% 抽取佣金
  5. 净额记入创作者钱包
  6. 生成结算单（商家侧 + 创作者侧）
  
佣金模型:
  初始: 平台抽佣 20% (行业标准)
  忠诚创作者: 累计收益越高，抽佣越低
    - < ¥10,000 → 20%
    - ¥10,000-50,000 → 15%
    - > ¥50,000 → 10%
```

---

## 4. 钱包与提现系统

### 4.1 资金流水

| 交易类型 | 说明 | 方向 |
|---------|------|------|
| 任务结算收入 | 完成任务获得的收益 | 入账 |
| 平台佣金 | 平台收取的服务费 | 出账 |
| 提现 | 提现到支付宝/微信/银行卡 | 出账 |
| 充值（商家）| 商家充值资金用于发布任务 | 入账 |
| 退款 | 任务取消/争议后退款 | 退账 |

### 4.2 提现规则

```yaml
min_withdrawal: ¥100         # 最低提现金额
max_withdrawal: ¥50,000      # 单次最高提现金额
cooling_period: 3 days       # 结算后冷静期才能提现
processing_time: 1-3 days    # 处理时间
channels:
  - alipay                   # 支付宝
  - wechat_pay               # 微信支付
  - bank_card                # 银行卡
fee:
  < ¥1,000: ¥5               # 小额提现手续费
  ≥ ¥1,000: free             # 免费
```

### 4.3 UI 设计

```yaml
页面: Wallet

布局:
  余额卡片:
    - 总余额（大字）
    - 冻结余额（结算中任务）
    - 可提现余额
    - 累计收益
    - 【提现】按钮
  
  交易记录:
    - 列表（时间 + 类型 + 金额 + 状态）
    - 筛选（全部/收入/提现）
    - 搜索（按时间范围）
  
  提现弹窗:
    - 提现金额输入
    - 收款方式选择（绑定支付宝/微信/银行卡）
    - 确认提现
    - 提现记录
```

---

## 5. 评价与信用体系

### 5.1 双端互评

```yaml
商家评价创作者:
  维度: 内容质量 / 交付及时性 / 配合度
  分数: 1-5 星
  文字评价（可选）

创作者评价商家:
  维度: 需求清晰度 / 结算速度 / 沟通体验
  分数: 1-5 星
  文字评价（可选）
```

### 5.2 信用分

```yaml
创作者信用分:
  初始: 100 分
  加分: 完成订单 +2, 优质内容被点赞 +1, 主动取消 -5
  扣分: 交付超时 -10, 内容违规 -20, 恶意刷量 -50
  低于 60 分: 限制接单

商家信用分:
  初始: 100 分
  加分: 按时结算 +2, 高预算任务 +1
  扣分: 无理驳回 -10, 延迟结算 -5, 恶意差评 -20
  低于 60 分: 限制发布任务
```

---

## 6. 数据报告

### 6.1 商家报告

```yaml
任务总览:
  - 发布任务数 / 进行中 / 已完成
  - 总投入预算 / 已结算金额
  - 平均 ROI

单任务报告:
  - 内容曝光量 / 互动量 / 转化量
  - 各创作者效果对比
  - 按天趋势图
  
创作者分析:
  - 效果 Top N 创作者
  - 各垂直领域表现
  - 粉丝画像重合度
```

### 6.2 创作者报告

```yaml
收益总览:
  - 本月收益 / 累计收益
  - 待结算金额 / 已提现金额
  - 平均任务单价

接单分析:
  - 接单数 / 完成率 / 通过率
  - 各平台任务分布
  - 各结算方式收益分布
```

---

## 7. 数据模型

### 7.1 核心表结构

```sql
-- 交易市场任务
CREATE TABLE marketplace_tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  advertiser_id UUID NOT NULL,           -- 商家 ID
  title VARCHAR(200) NOT NULL,
  description TEXT,
  platform VARCHAR(50)[] NOT NULL,
  settlement_type VARCHAR(10) NOT NULL CHECK (settlement_type IN ('CPS','CPE','CPM')),
  budget DECIMAL(12,2) NOT NULL,
  min_followers INT DEFAULT 0,
  min_engagement DECIMAL(5,2) DEFAULT 0,
  niche VARCHAR(100)[],
  content_type VARCHAR(50)[],
  guidelines TEXT,
  deadline TIMESTAMP,
  max_creators INT DEFAULT 1,
  status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open','in_progress','closed','cancelled')),
  commission_rate DECIMAL(5,2) DEFAULT 20.00,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- 接单记录
CREATE TABLE task_orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id UUID NOT NULL REFERENCES marketplace_tasks(id),
  creator_id UUID NOT NULL,
  status VARCHAR(20) DEFAULT 'accepted' CHECK (status IN ('accepted','creating','delivered','approved','rejected','settled','cancelled')),
  deliver_url TEXT,
  deliver_notes TEXT,
  approved_at TIMESTAMP,
  settled_amount DECIMAL(12,2),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- 钱包
CREATE TABLE wallets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE,
  balance DECIMAL(12,2) DEFAULT 0,
  frozen_balance DECIMAL(12,2) DEFAULT 0,
  total_earned DECIMAL(12,2) DEFAULT 0,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- 交易流水
CREATE TABLE wallet_transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  wallet_id UUID NOT NULL REFERENCES wallets(id),
  type VARCHAR(20) NOT NULL CHECK (type IN ('task_income','commission','withdrawal','deposit','refund')),
  amount DECIMAL(12,2) NOT NULL,
  balance_before DECIMAL(12,2) NOT NULL,
  balance_after DECIMAL(12,2) NOT NULL,
  reference_id UUID,                    -- 关联订单 ID
  status VARCHAR(20) DEFAULT 'completed',
  description TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);

-- 评价
CREATE TABLE reviews (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  order_id UUID NOT NULL REFERENCES task_orders(id),
  from_user_id UUID NOT NULL,
  to_user_id UUID NOT NULL,
  rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
  content TEXT,
  dimensions JSONB,                     -- 各维度评分
  created_at TIMESTAMP DEFAULT NOW()
);
```

---

## 8. 工作量预估

| 模块 | 后端（天）| 前端（天）| 测试（天）| 合计 |
|------|---------|---------|---------|------|
| 数据库设计 + 迁移 | 2 | - | 1 | 3 |
| 任务 CRUD API | 4 | - | 1 | 5 |
| 任务广场 UI | - | 3 | 1 | 4 |
| 接单/交付流程 | 3 | 2 | 2 | 7 |
| 结算引擎 | 5 | - | 3 | 8 |
| 钱包系统 | 3 | 2 | 2 | 7 |
| 评价体系 | 2 | 1 | 1 | 4 |
| 数据报告 | 3 | 3 | 2 | 8 |
| 提现处理 | 2 | 0.5 | 1 | 3.5 |
| 集成测试 + 文档 | 2 | 1 | 1 | 4 |
| **总计** | **26** | **12.5** | **15** | **53.5** |

---

## 9. 风险与应对

| 风险 | 影响 | 概率 | 应对 |
|------|------|------|------|
| 虚假数据刷量 | 商家损失 | 高 | 反作弊系统 + 人工抽查 + 数据源交叉验证 |
| 结算争议 | 信任危机 | 中 | 清晰的结算规则 + 争议仲裁流程 |
| 创作者/商家流失 | 市场冷启动 | 高 | 初期补贴策略 + 优质创作者定向邀请 |
| 支付合规（二清风险）| 法律风险 | 中 | 与持牌支付机构合作，不自建资金池 |
| 内容版权纠纷 | 法律纠纷 | 中 | 明确内容授权条款，平台免责声明 |

> **注：** 支付合规是最重要的法律风险。按照中国法规，平台不能直接沉淀资金。解决方案：接入持牌支付机构（如 Ping++、LianLian Global）的分账系统，资金不经过平台账户。
