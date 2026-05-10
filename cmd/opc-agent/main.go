package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wjames2000/opc-macs/internal/agent"
	"github.com/wjames2000/opc-macs/internal/config"
	"github.com/wjames2000/opc-macs/internal/hitl"
	"github.com/wjames2000/opc-macs/internal/memory"
	"github.com/wjames2000/opc-macs/internal/plugins"
	"github.com/wjames2000/opc-macs/internal/runtime"
)

var (
	version = "0.1.0"
	buildID = "dev"
)

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	showVersion := flag.Bool("version", false, "显示版本")
	flag.Parse()

	if *showVersion {
		fmt.Printf("OPC-Agent v%s (build %s)\n", version, buildID)
		os.Exit(0)
	}

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("配置加载失败：%v", err)
	}

	fmt.Printf("OPC-Agent v%s 启动中...\n", version)

	// 初始化结构化日志
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))
	slog.Info("system_start", "version", version, "build", buildID)

	// 初始化插件加载器
	pluginLoader := runtime.NewLoader(cfg.Runtime.PluginsDir)

	// 尝试从 .so 文件加载插件（Linux 生产模式）
	if err := pluginLoader.LoadAll(); err != nil {
		fmt.Printf("[系统] .so 插件目录不可用（%v），切换到内嵌模式\n", err)
	}

	if pluginLoader.Count() == 0 {
		fmt.Println("[系统] 使用内嵌 Agent 插件（开发模式）")
		if err := plugins.RegisterAll(pluginLoader); err != nil {
			log.Fatalf("内嵌插件注册失败：%v", err)
		}
	}
	fmt.Printf("[系统] 已加载 %d 个 Agent 插件\n", pluginLoader.Count())

	// 初始化模型客户端
	modelClient, err := agent.NewModelClient(cfg.Model.Provider, cfg.Model.APIBaseURL, cfg.Model.APIKey)
	if err != nil {
		log.Fatalf("模型初始化失败：%v\n请在 config.yaml 中配置正确的 model.provider 和 model.api_key", err)
	}
	fmt.Printf("[系统] 模型：%s → %s (%s)\n", cfg.Model.Provider, cfg.Model.Name, cfg.Model.APIBaseURL)

	// 初始化记忆引擎（带真实 embedding）
	embedder := memory.NewModelClientEmbedder(modelClient, cfg.Model.Name)
	memoryStore, err := memory.NewEmbeddedEngineWithEmbedder(cfg.Memory.StorePath, embedder)
	if err != nil {
		log.Printf("[警告] 记忆引擎初始化失败：%v，使用空引擎降级运行", err)
		memoryStore = memory.NewEmptyEngine()
	}
	defer memoryStore.Close()

	count, _ := memoryStore.Count()
	fmt.Printf("[系统] 记忆引擎就绪，已存储 %d 条记忆\n", count)

	// 初始化 HITL
	hitlHandler := hitl.NewHandler(os.Stdin, os.Stdout)

	// 初始化 Router
	router := agent.NewRouter(pluginLoader, cfg.Model.Name, modelClient)

	// 初始化 Reviewer
	reviewer := agent.NewReviewer(cfg.Model.Name, modelClient)

	// 信号处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\n\n正在退出...")
		memoryStore.Close()
		os.Exit(0)
	}()

	// 启动 REPL
	fmt.Println("\nOPC-Agent 已就绪，输入任务描述开始工作，输入 exit 退出。")
	printHelp()

	scanner := bufio.NewScanner(os.Stdin)
	session := agent.NewSession(50)
	tokenTracker := agent.NewTokenTracker()
	hasHistory := false

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}
		if input == "help" {
			printHelp()
			continue
		}
		if input == "plugins" {
			printPlugins(pluginLoader)
			continue
		}
		if input == "stats" {
			printStats(memoryStore)
			continue
		}
		if input == "usage" || input == "cost" {
			fmt.Println(tokenTracker.Summary())
			continue
		}
		if input == "audit" {
			fmt.Println(tokenTracker.RecentCalls(20))
			continue
		}
		if input == "new" || input == "reset" {
			session.Reset()
			hasHistory = false
			fmt.Println("[会话] 已重置，开始新的对话")
			continue
		}
		if input == "history" {
			fmt.Println(session.FormatHistory(50))
			continue
		}
		if input == "reload" {
			fmt.Println("[重载] 重新加载配置...")
			newCfg, err := config.Load(*configPath)
			if err != nil {
				fmt.Printf("[重载] 配置加载失败：%v\n", err)
				continue
			}
			// 重建模型客户端
			newClient, err := agent.NewModelClient(newCfg.Model.Provider, newCfg.Model.APIBaseURL, newCfg.Model.APIKey)
			if err != nil {
				fmt.Printf("[重载] 模型初始化失败：%v\n", err)
				continue
			}
			modelClient = newClient
			cfg = newCfg

			// 更新 embedder
			embedder = memory.NewModelClientEmbedder(modelClient, cfg.Model.Name)

			// 更新 Router 和 Reviewer
			router.UpdateModel(cfg.Model.Name, modelClient)
			reviewer = agent.NewReviewer(cfg.Model.Name, modelClient)

			fmt.Printf("[重载] 配置已更新：模型 %s → %s\n", cfg.Model.Provider, cfg.Model.Name)
			continue
		}

		if hasHistory {
			history := session.FormatHistory(5)
			_ = history
		}

		taskInput := input
		resultStr, routeName := processTask(context.Background(), input, pluginLoader, router, reviewer, memoryStore, hitlHandler, modelClient, embedder, cfg, session, tokenTracker)

		session.AddTurn(taskInput, routeName, resultStr)
		_ = hasHistory
	}

	fmt.Println("\n再见！")
}

func processTask(ctx context.Context, input string, loader *runtime.Loader,
	router *agent.Router, reviewer *agent.Reviewer,
	store memory.MemoryStore, hitlHandler *hitl.Handler,
	modelClient runtime.ModelClient, embedder memory.Embedder,
	cfg *config.Config, session *agent.Session,
	tracker *agent.TokenTracker) (resultStr string, routeName string) {

	startTime := time.Now()
	fmt.Printf("\n[任务] 处理中：%s\n", input)

	// 超时控制
	taskCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	slog.Info("task_start", "input", input)

	// 1. Router 分发
	routeResult, err := router.Route(taskCtx, input)
	if err != nil {
		fmt.Printf("[错误] 路由失败：%v\n", err)
		slog.Error("route_failed", "error", err, "input", input)
		return "", ""
	}
	if routeResult.Action == agent.RouteActionUnknown {
		fmt.Printf("%s\n", routeResult.Message)
		slog.Warn("route_unknown", "input", input)
		return "", ""
	}

	// 当使用 @agent_name 明确指定时，使用 @ 后面的内容作为任务输入
	taskInput := input
	if routeResult.Message != "" {
		taskInput = routeResult.Message
	}

	fmt.Printf("[路由] → %s\n", routeResult.Info.Name)
	if taskInput != input {
		fmt.Printf("[任务] 任务描述：%s\n", taskInput)
	}

	// 注入会话历史到 opts
	history := session.FormatHistory(5)
	_ = history

	// 2. 检索记忆
	memories, _ := store.Recall(taskCtx, taskInput, cfg.Memory.TopK)
	if len(memories) > 0 {
		fmt.Printf("[记忆] 找到 %d 条相关记忆\n", len(memories))
	}

	// 3. 调用插件 Execute
	opts := map[string]interface{}{
		"memories":      memories,
		"model_client":  modelClient,
		"model_name":    cfg.Model.Name,
		"history":       session.FormatHistory(5),
	}
	execResult, err := routeResult.Plugin.Execute(taskCtx, taskInput, opts)
	if err != nil {
		fmt.Printf("[错误] Agent 执行失败：%v\n", err)
		return "", ""
	}
	modelInfo := execResult.TokenUsage.ModelName
	if modelInfo == "" {
		modelInfo = routeResult.Info.ModelName
		if modelInfo == "" {
			modelInfo = cfg.Model.Name
		}
	}
	fmt.Printf("[执行] 完成 (model=%s in=%d out=%d)\n",
		modelInfo, execResult.TokenUsage.InputTokens, execResult.TokenUsage.OutputTokens)

	tracker.RecordCall(routeResult.Info.Name, modelInfo, "execute",
		execResult.TokenUsage.InputTokens, execResult.TokenUsage.OutputTokens,
		time.Since(startTime), err == nil)

	// 输出 Agent 思考过程
	if execResult.RawTrace != "" {
		fmt.Println("\n┌─── Agent 思考过程 ────────────────────────────")
		fmt.Println(execResult.RawTrace)
		fmt.Println("└──────────────────────────────────────────────────")
	}

	// 4. Reviewer 审查
	reviewResult, _ := reviewer.Review(taskCtx, execResult.Data,
		extractCheckpoints(routeResult.Info), execResult.RawTrace)

	// 输出审查思考过程
	if reviewResult != nil && reviewResult.Trace != "" {
		fmt.Println("\n┌─── 审查思考过程 ───────────────────────────────")
		fmt.Println(reviewResult.Trace)
		fmt.Println("└──────────────────────────────────────────────────")
	}

	if reviewResult != nil && !reviewResult.Passed {
		fmt.Printf("[审查] 未通过 (评分 %.1f/5.0)：%s\n",
			reviewResult.Score, reviewResult.Summary)

		if reviewResult.ShouldRetry {
			fmt.Println("[重试] 正在重试...")
			execResult, err = routeResult.Plugin.Execute(taskCtx, input, opts)
			if err == nil {
				reviewResult, _ = reviewer.Review(taskCtx, execResult.Data,
					extractCheckpoints(routeResult.Info), execResult.RawTrace)
				if reviewResult != nil && reviewResult.Passed {
					fmt.Println("[审查] 重试后通过")
				}
			}
		}
	} else {
		fmt.Printf("[审查] 通过 (评分 %.1f/5.0)\n", reviewResult.Score)
	}

	// 5. HITL 检查
	if routeResult.Info.RequiresHITL {
		op := hitl.Operation{
			Type:        routeResult.Info.Name,
			Description: fmt.Sprintf("Agent '%s' 请求执行操作", routeResult.Info.Name),
		}
		approved, err := hitlHandler.Confirm(taskCtx, op)
		if err != nil {
			fmt.Printf("[HITL] 错误：%v\n", err)
			return
		}
		if !approved {
			fmt.Println("[HITL] 操作已取消")
			return
		}
		fmt.Println("[HITL] 已确认")
	}

	// 6. 写入记忆
	keyDecisions := extractKeyDecisions(execResult.Data)
	entry := memory.BuildMemoryEntryWithEmbedder(taskCtx,
		routeResult.Info.Name,
		taskInput,
		fmt.Sprintf("%+v", execResult.Data),
		keyDecisions,
		map[string]string{
			"model":      modelInfo,
			"tokens_in":  fmt.Sprintf("%d", execResult.TokenUsage.InputTokens),
			"tokens_out": fmt.Sprintf("%d", execResult.TokenUsage.OutputTokens),
		},
		embedder,
	)
	if err := store.Store(taskCtx, entry); err != nil {
		fmt.Printf("[警告] 记忆写入失败：%v\n", err)
	} else {
		fmt.Println("[记忆] 已存储")
	}

	// 7. 输出结果
	elapsed := time.Since(startTime)
	resultStr = fmt.Sprintf("%+v", execResult.Data)
	fmt.Printf("\n══════════ 输出结果 (%.2fs) ══════════\n", elapsed.Seconds())
	fmt.Println(resultStr)
	fmt.Printf("══════════════════════════════════════\n")

	slog.Info("task_complete",
		"agent", routeResult.Info.Name,
		"model", modelInfo,
		"tokens_in", execResult.TokenUsage.InputTokens,
		"tokens_out", execResult.TokenUsage.OutputTokens,
		"duration_ms", elapsed.Milliseconds(),
	)

	return
}

func printHelp() {
	fmt.Println("\n可用命令：")
	fmt.Println("  @<Agent名> <任务>  指定 Agent 处理（如 @copywriter 写文案）")
	fmt.Println("  #<技能> <任务>     指定技能处理（如 #文案 推广文案、#邮件 投诉信）")
	fmt.Println("  <自然语言>          由 Router 自动识别意图并分发")
	fmt.Println("  plugins             查看已加载的 Agent 插件")
	fmt.Println("  usage/cost         查看 Token 用量与费用估算")
	fmt.Println("  audit              查看最近调用记录")
	fmt.Println("  stats               查看系统统计信息")
	fmt.Println("  new                 开始新会话（清空上下文）")
	fmt.Println("  history             查看当前会话历史")
	fmt.Println("  reload              重新加载配置（模型/API 等）")
	fmt.Println("  help                显示帮助")
	fmt.Println("  exit                退出")
}

func printPlugins(loader *runtime.Loader) {
	plugins := loader.List()
	if len(plugins) == 0 {
		fmt.Println("没有已加载的 Agent 插件")
		return
	}
	fmt.Printf("\n已加载的 Agent 插件（%d 个）：\n", len(plugins))
	for _, p := range plugins {
		fmt.Printf("  - %s v%s: %s\n", p.Name, p.Version, p.Summary)
	}
}

func printStats(store memory.MemoryStore) {
	count, _ := store.Count()
	fmt.Printf("\n系统统计：\n")
	fmt.Printf("  记忆条目数：%d\n", count)
}

func extractCheckpoints(info runtime.PluginInfo) []string {
	return info.Tags
}

func extractKeyDecisions(data interface{}) []string {
	if m, ok := data.(map[string]interface{}); ok {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	}
	return nil
}
