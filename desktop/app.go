package main

import (
	"context"
	"fmt"

	"github.com/wjames2000/opc-macs/internal/plugins"
	"github.com/wjames2000/opc-macs/internal/runtime"
)

type App struct {
	ctx    context.Context
	loader *runtime.Loader
}

func NewApp() *App {
	loader := runtime.NewLoader("")
	if err := plugins.RegisterAll(loader); err != nil {
		fmt.Printf("[Wails] 插件注册失败：%v\n", err)
	}
	return &App{loader: loader}
}

func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	fmt.Printf("[Wails] OPC-Agent 桌面端启动（%d 个 Agent）\n", a.loader.Count())
}

func (a *App) OnShutdown(ctx context.Context) {
	fmt.Println("[Wails] 桌面端关闭")
}

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

	result, err := plugin.Execute(a.ctx, req.Input, map[string]interface{}{})
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
