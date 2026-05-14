package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wjames2000/opc-macs/internal/agent"
	"github.com/wjames2000/opc-macs/internal/config"
	"github.com/wjames2000/opc-macs/internal/plugins"
	"github.com/wjames2000/opc-macs/internal/runtime"
)

type UserSettings struct {
	APIBaseURL   string `json:"api_base_url"`
	APIKey       string `json:"api_key"`
	Provider     string `json:"provider"`
	ModelName    string `json:"model_name"`
	DefaultAgent string `json:"default_agent"`
}

type App struct {
	ctx          context.Context
	loader       *runtime.Loader
	modelClient  runtime.ModelClient
	cfg          *config.Config
	settings     *UserSettings
	settingsPath string
	tracker      *agent.TokenTracker
	activityLog  []ActivityLogEntry
}

type ActivityLogEntry struct {
	Agent  string `json:"agent"`
	Input  string `json:"input"`
	Output string `json:"output"`
	Time   string `json:"time"`
	Tokens int    `json:"tokens"`
}

type DashboardStats struct {
	AgentCount     int               `json:"agent_count"`
	TotalCalls     int               `json:"total_calls"`
	TodayCalls     int               `json:"today_calls"`
	TodayTokens    int               `json:"today_tokens"`
	TotalCost      string            `json:"total_cost"`
	RecentActivity []ActivityLogEntry `json:"recent_activity"`
}

type ConversationSummary struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Preview string `json:"preview"`
	Turns   int    `json:"turns"`
	Time    string `json:"time"`
}

type AgentInfo struct {
	Name         string `json:"name"`
	Summary      string `json:"summary"`
	Version      string `json:"version"`
	Model        string `json:"model"`
	RequiresHITL bool   `json:"requires_hitl"`
	Tags         string `json:"tags"`
}

func settingsFilePath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "opc-agent")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "settings.json")
}

func NewApp() *App {
	loader := runtime.NewLoader("")
	plugins.RegisterAll(loader)

	settingsPath := settingsFilePath()
	settings := &UserSettings{
		APIBaseURL:   "",
		APIKey:       "",
		Provider:     "deepseek",
		ModelName:    "deepseek-v4-flash",
		DefaultAgent: "copywriter",
	}

	// Load saved settings
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		json.Unmarshal(data, settings)
	}

	// Init tracker
	tracker := agent.NewTokenTracker()

	// Create model client from settings
	client := createModelClient(settings)

	fmt.Printf("[Wails] 已加载 %d 个 Agent\n", loader.Count())
	return &App{
		loader:       loader,
		modelClient:  client,
		settings:     settings,
		settingsPath: settingsPath,
		tracker:      tracker,
		activityLog:  make([]ActivityLogEntry, 0),
	}
}

func createModelClient(s *UserSettings) runtime.ModelClient {
	if s.APIKey == "" {
		return nil
	}
	baseURL := s.APIBaseURL
	client, err := agent.NewModelClient(s.Provider, baseURL, s.APIKey)
	if err != nil {
		fmt.Printf("[Wails] 模型客户端创建失败：%v\n", err)
		return nil
	}
	return client
}

func (a *App) OnStartup(ctx context.Context) { a.ctx = ctx }
func (a *App) OnShutdown(ctx context.Context) {}

// ===== Settings =====

func (a *App) GetSettings() *UserSettings {
	return a.settings
}

func (a *App) SaveSettings(s UserSettings) bool {
	savePath := a.settingsPath
	if savePath == "" {
		savePath = settingsFilePath()
	}

	data, _ := json.MarshalIndent(s, "", "  ")
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		fmt.Printf("[Wails] 保存设置失败：%v\n", err)
		return false
	}

	// Recreate model client with new settings
	a.settings = &s
	a.modelClient = createModelClient(&s)
	return true
}

// ===== Agents =====

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

// ===== Execute =====

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

	if a.modelClient == nil {
		return ExecuteResponse{
			Success: false,
			Error:   "未配置模型 API。请在设置中配置 API Key。",
		}
	}

	opts := map[string]interface{}{
		"model_client": a.modelClient,
		"model_name":   a.settings.ModelName,
	}

	result, err := plugin.Execute(a.ctx, req.Input, opts)
	if err != nil {
		a.tracker.RecordCall(req.Agent, a.settings.ModelName, "execute", 0, 0, 0, false)
		return ExecuteResponse{Success: false, Error: err.Error()}
	}

	inTokens := result.TokenUsage.InputTokens
	outTokens := result.TokenUsage.OutputTokens
	a.tracker.RecordCall(req.Agent, a.settings.ModelName, "execute", inTokens, outTokens, 0, true)

	// Log activity
	var outputStr string
	if data, ok := result.Data.(map[string]interface{}); ok {
		if s, ok := data["summary"]; ok {
			outputStr = fmt.Sprintf("%v", s)
		} else if s, ok := data["short_copy"]; ok {
			outputStr = fmt.Sprintf("%v", s)
		} else if s, ok := data["title"]; ok {
			outputStr = fmt.Sprintf("%v", s)
		} else {
			outputStr = fmt.Sprintf("%v", result.Data)
		}
	} else {
		outputStr = fmt.Sprintf("%v", result.Data)
	}
	if len(outputStr) > 80 {
		outputStr = outputStr[:80] + "..."
	}

	a.activityLog = append(a.activityLog, ActivityLogEntry{
		Agent:  req.Agent,
		Input:  req.Input,
		Output: outputStr,
		Time:   time.Now().Format("15:04:05"),
		Tokens: inTokens + outTokens,
	})

	return ExecuteResponse{
		Success: true,
		Data:    result.Data,
		Tokens:  TokenInfo{Input: inTokens, Output: outTokens},
	}
}

func (a *App) GetVersion() map[string]string {
	return map[string]string{
		"version": "0.1.0",
		"agents":  fmt.Sprintf("%d", a.loader.Count()),
	}
}

// ===== Dashboard =====

func (a *App) GetDashboardStats() DashboardStats {
	recent := a.activityLog
	if len(recent) > 10 {
		recent = recent[len(recent)-10:]
	}

	today := a.tracker.TodayStats()
	usage := a.tracker.AgentUsage()
	totalCalls := 0
	totalCost := 0.0
	for _, u := range usage {
		totalCalls += u.Calls
		totalCost += u.Cost
	}

	return DashboardStats{
		AgentCount:     a.loader.Count(),
		TotalCalls:     totalCalls,
		TodayCalls:     today.Calls,
		TodayTokens:    today.Tokens,
		TotalCost:      fmt.Sprintf("$%.4f", totalCost),
		RecentActivity: recent,
	}
}

// ===== Usage =====

func (a *App) GetUsageStats() []agent.AgentUsageItem {
	return a.tracker.AgentUsage()
}

// ===== Agent List (full info) =====

func (a *App) GetAgentList() []AgentInfo {
	list := a.loader.List()
	result := make([]AgentInfo, 0, len(list))
	for _, p := range list {
		info := AgentInfo{
			Name:         p.Name,
			Summary:      p.Summary,
			Version:      p.Version,
			Model:        p.ModelName,
			RequiresHITL: p.RequiresHITL,
		}
		if len(p.Tags) > 0 {
			info.Tags = strings.Join(p.Tags, ", ")
		}
		result = append(result, info)
	}
	return result
}

