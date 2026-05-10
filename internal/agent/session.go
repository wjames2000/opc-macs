package agent

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type Turn struct {
	Input    string
	Agent    string
	Output   string
	Time     time.Time
}

type Session struct {
	mu     sync.RWMutex
	Turns  []Turn
	MaxLen int
}

func NewSession(maxLen int) *Session {
	return &Session{
		Turns:  make([]Turn, 0),
		MaxLen: maxLen,
	}
}

func (s *Session) AddTurn(input, agent, output string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Turns = append(s.Turns, Turn{
		Input:  input,
		Agent:  agent,
		Output: output,
		Time:   time.Now(),
	})
	if len(s.Turns) > s.MaxLen {
		overflow := len(s.Turns) - s.MaxLen
		s.Turns = s.Turns[overflow:]
	}
}

func (s *Session) History() []Turn {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Turn, len(s.Turns))
	copy(result, s.Turns)
	return result
}

func (s *Session) FormatHistory(maxTurns int) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.Turns) == 0 {
		return ""
	}
	start := 0
	if len(s.Turns) > maxTurns {
		start = len(s.Turns) - maxTurns
	}
	var sb strings.Builder
	sb.WriteString("以下是本次会话的历史对话记录，供参考：\n")
	for i := start; i < len(s.Turns); i++ {
		t := s.Turns[i]
		sb.WriteString(fmt.Sprintf("--- 第 %d 轮 ---\n", i+1))
		sb.WriteString(fmt.Sprintf("用户：%s\n", t.Input))
		sb.WriteString(fmt.Sprintf("Agent (%s)：%s\n", t.Agent, truncateStr(t.Output, 80)))
	}
	return sb.String()
}

func (s *Session) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Turns = make([]Turn, 0)
}

func (s *Session) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Turns)
}

func truncateStr(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
