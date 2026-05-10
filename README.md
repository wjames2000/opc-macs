# OPC-Agent

多任务智能协助系统 — 为 OPC（一人公司）创业者打造的 AI Agent 管理平台。

## 架构

```
┌─ REPL 交互 ─────────────────────────────┐
│  Router → Plugin.Execute → Reviewer →   │
│  HITL → Memory.Store → Output          │
└─────────────────────────────────────────┘
         │ 插件架构（plugin.Open）
    ┌────┼────┐
    ▼    ▼    ▼
 copywriter email_sorter xhs_poster
  (.so)     (.so)        (.so)
```

- **插件架构**：业务 Agent 编译为 Go Plugin (.so)，主程序运行时不依赖具体插件代码
- **Harness 控制**：角色边界、状态机、产物契约、护栏规则四重约束
- **向量记忆**：嵌入式引擎 + cosine similarity + 文件持久化
- **HITL 确认**：敏感操作终端交互确认

## 快速开始

```bash
# 前提：Go 1.23+
git clone <repo>
cd opc-agent

# 编译主程序
make build-main

# 编译所有插件
make build-plugins

# 运行
make run
```

进入 REPL 后输入自然语言任务描述，例如：

```
> 帮我写个智能水杯的文案
> 处理这封邮件：产品坏了我要退款
> 帮我写一篇小红书笔记推广这个产品
```

## 可用命令

```
<自然语言>  输入任务描述
plugins     查看已加载的 Agent
stats       查看统计信息
help        显示帮助
exit        退出
```

## 插件开发

创建新 Agent 只需三步：

1. 在 `plugins/new_agent/` 目录下创建 `plugin.go` 实现 `AgentPlugin` 接口
2. 添加 `SKILL.md` 定义技能行为
3. 编译：`go build -buildmode=plugin -o build/plugins/new_agent.so .`

## 项目结构

```
├── cmd/opc-agent/           # 主程序入口
├── internal/
│   ├── config/              # 配置加载
│   ├── harness/             # Harness 控制
│   ├── memory/              # 向量记忆引擎
│   ├── hitl/                # HITL 确认
│   ├── runtime/             # 插件运行时
│   └── agent/               # Router + Reviewer
├── plugins/                 # Agent 插件源码
│   ├── copywriter/          # 文案生成
│   ├── email_sorter/        # 邮件分类
│   └── xhs_poster/          # 小红书内容
├── pkg/contracts/           # JSON Schema
├── docs/                    # 项目文档
└── build/                   # 构建产物
```
