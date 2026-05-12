package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wjames2000/opc-macs/internal/agent"
	"github.com/wjames2000/opc-macs/internal/config"
	"github.com/wjames2000/opc-macs/internal/plugins"
	"github.com/wjames2000/opc-macs/internal/runtime"
)

type App struct {
	ctx         context.Context
	loader      *runtime.Loader
	modelClient runtime.ModelClient
	cfg         *config.Config
}

func NewApp() *App {
	loader := runtime.NewLoader("")
	if err := plugins.RegisterAll(loader); err != nil {
		fmt.Printf("[Wails] 插件注册失败：%v\n", err)
	}

	// 尝试加载配置
	var client runtime.ModelClient
	var cfg *config.Config

	if data, err := os.ReadFile("../config.dev.yaml"); err == nil {
		if c, err := config.Load("../config.dev.yaml"); err == nil {
			cfg = c
			client, _ = agent.NewModelClient(cfg.Model.Provider, cfg.Model.APIBaseURL, cfg.Model.APIKey)
		}
		_ = data
	}

	if client == nil {
		fmt.Println("[Wails] 未找到 config.yaml，使用本地模式（需连接 SaaS 后端）")
	}

	fmt.Printf("[Wails] 已加载 %d 个 Agent\n", loader.Count())
	return &App{loader: loader, modelClient: client, cfg: cfg}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) OnShutdown(ctx context.Context) {}

func (a *App) GetAgents() []map[string]interface{} {
	var result []map[string]interface{}
	for _, p := range a.loader.List() {
		result = append(result, map[string]interface{}{
			"name":    p.Name,
			"summary": p.Summary,
			"version": p.Version,
			"model":   p.ModelName,
			"hitl":    p.RequiresHITL,
		})
	}
	return result
}

type ExecuteRequest struct {
	Agent string `json:"agent"`
	Input string `json:"input"`
}

type ExecuteResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Tokens  TokenInfo   `json:"tokens"`
}

type TokenInfo struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

func (a *App) Execute(req ExecuteRequest) ExecuteResponse {
	plugin, ok := a.loader.Get(req.Agent)
	if !ok {
		return ExecuteResponse{
			Success: false,
			Error:   fmt.Sprintf("Agent '%s' 未找到", req.Agent),
		}
	}

	opts := map[string]interface{}{}
	if a.modelClient != nil {
		opts["model_client"] = a.modelClient
		opts["model_name"] = a.cfg.Model.Name
	} else {
		return ExecuteResponse{
			Success: false,
			Error:   "未配置模型 API。请在 config.yaml 中设置 api_key，或连接 SaaS 后端。",
		}
	}

	result, err := plugin.Execute(a.ctx, req.Input, opts)
	if err != nil {
		return ExecuteResponse{
			Success: false,
			Error:   err.Error(),
		}
	}

	return ExecuteResponse{
		Success: true,
		Data:    result.Data,
		Tokens: TokenInfo{
			Input:  result.TokenUsage.InputTokens,
			Output: result.TokenUsage.OutputTokens,
		},
	}
}

func (a *App) GetVersion() map[string]string {
	return map[string]string{
		"version": "0.1.0",
		"agents":  fmt.Sprintf("%d", a.loader.Count()),
	}
}
