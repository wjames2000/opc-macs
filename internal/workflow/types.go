package workflow

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Workflow struct {
	Name        string  `yaml:"name"`
	Description string  `yaml:"description"`
	Version     string  `yaml:"version"`
	Trigger     string  `yaml:"trigger"`
	Schedule    string  `yaml:"schedule,omitempty"`
	Steps       []*Step `yaml:"steps"`
}

type Step struct {
	ID        string         `yaml:"id"`
	Agent     string         `yaml:"agent,omitempty"`
	Input     string         `yaml:"input"`
	Output    string         `yaml:"output"`
	Parallel  []*ParallelStep `yaml:"parallel,omitempty"`
	Aggregate string         `yaml:"aggregate,omitempty"`
	Condition string         `yaml:"condition,omitempty"`
	Retry     *RetryConfig   `yaml:"retry,omitempty"`
}

type ParallelStep struct {
	ID     string `yaml:"id"`
	Agent  string `yaml:"agent"`
	Input  string `yaml:"input"`
	Output string `yaml:"output"`
}

type RetryConfig struct {
	MaxAttempts int `yaml:"max_attempts"`
	DelayMs     int `yaml:"delay_ms"`
}

type StepLog struct {
	StepID    string
	Agent     string
	Status    string
	StartTime time.Time
	Duration  time.Duration
	Input     string
	Output    string
	Error     string
}

type Session struct {
	ID        string
	Workflow  *Workflow
	Status    string
	StartTime time.Time
	EndTime   time.Time
	Context   map[string]interface{} // step output variables
	Logs      []StepLog
	mu        sync.RWMutex
}

func (s *Session) SetVar(name string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Context[name] = value
}

func (s *Session) GetVar(name string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Context[name]
	return v, ok
}

func (s *Session) AddLog(log StepLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Logs = append(s.Logs, log)
}

// FormatStatus 格式化工作流执行状态用于显示
func (s *Session) FormatStatus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📊 工作流：%s (%s)\n", s.Workflow.Name, s.Status))
	sb.WriteString(fmt.Sprintf("   开始：%s\n", s.StartTime.Format("15:04:05")))
	if !s.EndTime.IsZero() {
		sb.WriteString(fmt.Sprintf("   结束：%s (%.1fs)\n", s.EndTime.Format("15:04:05"), s.EndTime.Sub(s.StartTime).Seconds()))
	}
	sb.WriteString("\n")
	for _, log := range s.Logs {
		icon := "⏳"
		switch log.Status {
		case "completed":
			icon = "✅"
		case "running":
			icon = "🔄"
		case "failed":
			icon = "❌"
		case "skipped":
			icon = "⏭️"
		}
		agent := log.Agent
		if agent == "" {
			agent = "parallel"
		}
		sb.WriteString(fmt.Sprintf("  %s %-20s %-10s", icon, log.StepID, log.Status))
		if log.Duration > 0 {
			sb.WriteString(fmt.Sprintf(" %.1fs", log.Duration.Seconds()))
		}
		if log.Error != "" {
			sb.WriteString(fmt.Sprintf(" %s", log.Error))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// FormatOutput 格式化工作流最终输出
func (s *Session) FormatOutput() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n══════════ 工作流完成：%s ══════════\n", s.Workflow.Name))
	sb.WriteString(fmt.Sprintf("状态：%s | 耗时：%.1fs\n", s.Status, s.EndTime.Sub(s.StartTime).Seconds()))
	
	// 显示每个步骤的输出
	for _, log := range s.Logs {
		if log.Status == "completed" && log.Output != "" {
			sb.WriteString(fmt.Sprintf("\n--- %s (%s) ---\n", log.StepID, log.Agent))
			sb.WriteString(fmt.Sprintf("%s\n", truncateStr(log.Output, 200)))
		}
	}
	sb.WriteString(fmt.Sprintf("\n══════════════════════════════════════\n"))
	return sb.String()
}

func truncateStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
