package agent

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type CallRecord struct {
	Time           time.Time
	Agent          string
	Model          string
	InputTokens    int
	OutputTokens   int
	Duration       time.Duration
	Success        bool
	PluginCallType string // "execute" | "review" | "classify"
}

type TokenTracker struct {
	mu     sync.RWMutex
	calls  []CallRecord
	models map[string]float64 // model name → $ per 1K input tokens
}

func NewTokenTracker() *TokenTracker {
	return &TokenTracker{
		calls: make([]CallRecord, 0),
		models: map[string]float64{
			"gpt-4o":            0.0025,
			"gpt-4o-mini":       0.00015,
			"deepseek-chat":     0.00027,
			"deepseek-v4-flash": 0.00027,
			"deepseek-v4-pro":   0.0014,
			"claude-3-5-sonnet": 0.003,
			"qwen-turbo":        0.0005,
			"moonshot-v1-8k":    0.0012,
			"gemini-2.0-flash":  0.0001,
			"text-embedding-3-small": 0.00002,
		},
	}
}

func (t *TokenTracker) Record(call CallRecord) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.calls = append(t.calls, call)
}

func (t *TokenTracker) Summary() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var totalIn, totalOut int
	agentTokens := make(map[string][2]int) // agent → [in, out]
	agentCalls := make(map[string]int)

	for _, c := range t.calls {
		totalIn += c.InputTokens
		totalOut += c.OutputTokens
		key := c.Agent
		if key == "" {
			key = "system"
		}
		v := agentTokens[key]
		agentTokens[key] = [2]int{v[0] + c.InputTokens, v[1] + c.OutputTokens}
		agentCalls[key]++
	}

	var sb strings.Builder
	sb.WriteString("\n┌─── Token 用量统计 ────────────────────────\n")
	sb.WriteString(fmt.Sprintf("│ 总调用次数：%d\n", len(t.calls)))
	sb.WriteString(fmt.Sprintf("│ 总输入 Token：%d (≈ $%.4f)\n", totalIn, t.estimateCost("", totalIn, 0)))
	sb.WriteString(fmt.Sprintf("│ 总输出 Token：%d (≈ $%.4f)\n", totalOut, t.estimateCost("", 0, totalOut)))
	sb.WriteString(fmt.Sprintf("│ 总费用估算：$%.4f\n", t.estimateCost("", totalIn, totalOut)))
	sb.WriteString("│\n│ 按 Agent 统计：\n")

	for agent, tokens := range agentTokens {
		inEst := t.estimateCost("", tokens[0], 0)
		outEst := t.estimateCost("", 0, tokens[1])
		sb.WriteString(fmt.Sprintf("│   %-15s %3d次  in=%6d  out=%6d  ≈$%.4f\n",
			agent, agentCalls[agent], tokens[0], tokens[1], inEst+outEst))
	}
	sb.WriteString("└──────────────────────────────────────────\n")
	return sb.String()
}

// ---- Structured data for desktop/web UI ----

type AgentUsageItem struct {
	Agent     string  `json:"agent"`
	Model     string  `json:"model"`
	Calls     int     `json:"calls"`
	InTokens  int     `json:"in_tokens"`
	OutTokens int     `json:"out_tokens"`
	Cost      float64 `json:"cost"`
}

type ActivityItem struct {
	Agent       string `json:"agent"`
	InputTokens int    `json:"input_tokens"`
	OutputTokens int   `json:"output_tokens"`
	Time        string `json:"time"`
	Success     bool   `json:"success"`
}

type TodayStats struct {
	Calls   int `json:"calls"`
	Tokens  int `json:"tokens"`
}

func (t *TokenTracker) AgentUsage() []AgentUsageItem {
	t.mu.RLock()
	defer t.mu.RUnlock()

	agentMap := make(map[string]*AgentUsageItem)
	agentModel := make(map[string]string)

	for _, c := range t.calls {
		name := c.Agent
		if name == "" {
			name = "system"
		}
		if _, ok := agentMap[name]; !ok {
			agentMap[name] = &AgentUsageItem{Agent: name}
		}
		agentMap[name].Calls++
		agentMap[name].InTokens += c.InputTokens
		agentMap[name].OutTokens += c.OutputTokens
		if c.Model != "" {
			agentModel[name] = c.Model
		}
	}

	result := make([]AgentUsageItem, 0, len(agentMap))
	for _, item := range agentMap {
		item.Model = agentModel[item.Agent]
		item.Cost = t.estimateCost(item.Model, item.InTokens, item.OutTokens)
		result = append(result, *item)
	}
	return result
}

func (t *TokenTracker) RecentActivity(n int) []ActivityItem {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if n <= 0 || len(t.calls) == 0 {
		return []ActivityItem{}
	}
	start := len(t.calls) - n
	if start < 0 {
		start = 0
	}
	result := make([]ActivityItem, 0, n)
	for i := start; i < len(t.calls); i++ {
		c := t.calls[i]
		result = append(result, ActivityItem{
			Agent:        c.Agent,
			InputTokens:  c.InputTokens,
			OutputTokens: c.OutputTokens,
			Time:         c.Time.Format("15:04:05"),
			Success:      c.Success,
		})
	}
	return result
}

func (t *TokenTracker) TodayStats() TodayStats {
	t.mu.RLock()
	defer t.mu.RUnlock()

	now := time.Now()
	var stats TodayStats
	for _, c := range t.calls {
		if c.Time.Year() == now.Year() && c.Time.YearDay() == now.YearDay() {
			stats.Calls++
			stats.Tokens += c.InputTokens + c.OutputTokens
		}
	}
	return stats
}

func (t *TokenTracker) RecentCalls(n int) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.calls) == 0 {
		return "暂无调用记录"
	}
	if n > len(t.calls) {
		n = len(t.calls)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n最近 %d 次调用：\n", n))
	start := len(t.calls) - n
	for i := start; i < len(t.calls); i++ {
		c := t.calls[i]
		status := "✅"
		if !c.Success {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("  %s [%s] %s → %s (in=%d out=%d %.2fs)\n",
			status, c.Time.Format("15:04:05"), c.Agent, c.Model,
			c.InputTokens, c.OutputTokens, c.Duration.Seconds()))
	}
	return sb.String()
}

func (t *TokenTracker) estimateCost(model string, inTokens, outTokens int) float64 {
	rate, ok := t.models[model]
	if !ok || model == "" {
		rate = 0.001 // default fallback
	}
	costIn := float64(inTokens) / 1000.0 * rate
	costOut := float64(outTokens) / 1000.0 * rate * 2 // output costs ~2x
	if outTokens == 0 && inTokens > 0 {
		return float64(inTokens) / 1000.0 * rate
	}
	return costIn + costOut
}

func (t *TokenTracker) RecordCall(agent, model, callType string, inTokens, outTokens int, duration time.Duration, success bool) {
	call := CallRecord{
		Time:           time.Now(),
		Agent:          agent,
		Model:          model,
		InputTokens:    inTokens,
		OutputTokens:   outTokens,
		Duration:       duration,
		Success:        success,
		PluginCallType: callType,
	}
	t.Record(call)
}
