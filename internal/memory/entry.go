package memory

import "time"

type MemoryEntry struct {
	ID           string            `json:"id"`
	CreatedAt    time.Time         `json:"created_at"`
	TaskType     string            `json:"task_type"`
	UserInput    string            `json:"user_input"`
	AgentOutput  string            `json:"agent_output"`
	KeyDecisions []string          `json:"key_decisions"`
	Embedding    []float32         `json:"embedding"`
	Metadata     map[string]string `json:"metadata"`
}
