package memory

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wjames2000/opc-macs/internal/runtime"
)

func CosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}

type EmbeddedEngine struct {
	mu        sync.RWMutex
	entries   []*MemoryEntry
	storePath string
	embedder  Embedder // 可选：真实 embedding 模型
}

// Embedder 抽象 embedding 生成接口
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

func NewEmbeddedEngine(storePath string) (*EmbeddedEngine, error) {
	return NewEmbeddedEngineWithEmbedder(storePath, nil)
}

func NewEmbeddedEngineWithEmbedder(storePath string, embedder Embedder) (*EmbeddedEngine, error) {
	dir := filepath.Dir(storePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("memory: create dir failed: %w", err)
	}
	engine := &EmbeddedEngine{
		entries:   make([]*MemoryEntry, 0),
		storePath: storePath,
		embedder:  embedder,
	}
	if err := engine.loadFromDisk(); err != nil {
		return engine, nil
	}
	return engine, nil
}

func NewEmptyEngine() *EmbeddedEngine {
	return &EmbeddedEngine{
		entries: make([]*MemoryEntry, 0),
	}
}

// ModelClientEmbedder wraps a runtime.ModelClient as an Embedder
type ModelClientEmbedder struct {
	client runtime.ModelClient
	model  string
}

func NewModelClientEmbedder(client runtime.ModelClient, model string) *ModelClientEmbedder {
	return &ModelClientEmbedder{client: client, model: model}
}

func (m *ModelClientEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	resp, err := m.client.Embed(ctx, runtime.EmbedRequest{
		Model: m.model,
		Input: text,
	})
	if err != nil {
		return nil, err
	}
	return resp.Embedding, nil
}

func (e *EmbeddedEngine) Store(ctx context.Context, entry *MemoryEntry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.entries = append(e.entries, entry)
	return e.atomicPersist()
}

func (e *EmbeddedEngine) atomicPersist() error {
	if e.storePath == "" {
		return nil
	}

	tmpPath := e.storePath + ".tmp"
	data, err := json.MarshalIndent(e.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("memory: marshal failed: %w", err)
	}

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("memory: write tmp failed: %w", err)
	}

	if err := os.Rename(tmpPath, e.storePath); err != nil {
		return fmt.Errorf("memory: rename failed: %w", err)
	}

	return nil
}

type scoredEntry struct {
	entry *MemoryEntry
	score float32
}

func (e *EmbeddedEngine) Search(ctx context.Context, query []float32, topK int) ([]*MemoryEntry, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if topK <= 0 {
		topK = 5
	}

	var candidates []scoredEntry
	for _, entry := range e.entries {
		score := CosineSimilarity(query, entry.Embedding)
		if score >= 0.7 {
			candidates = append(candidates, scoredEntry{entry, score})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if len(candidates) > topK {
		candidates = candidates[:topK]
	}

	result := make([]*MemoryEntry, len(candidates))
	for i, c := range candidates {
		result[i] = c.entry
	}
	return result, nil
}

func (e *EmbeddedEngine) Recall(ctx context.Context, text string, topK int) ([]*MemoryEntry, error) {
	var embedding []float32
	if e.embedder != nil {
		var err error
		embedding, err = e.embedder.Embed(ctx, text)
		if err != nil {
			// 回退到 mock embedding
			embedding = mockEmbedding(text)
		}
	} else {
		embedding = mockEmbedding(text)
	}
	return e.Search(ctx, embedding, topK)
}

func (e *EmbeddedEngine) Count() (int, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.entries), nil
}

func (e *EmbeddedEngine) Close() error {
	return nil
}

func (e *EmbeddedEngine) loadFromDisk() error {
	data, err := os.ReadFile(e.storePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &e.entries); err != nil {
		backupPath := e.storePath + ".bak"
		if backupData, err2 := os.ReadFile(backupPath); err2 == nil {
			if err2 := json.Unmarshal(backupData, &e.entries); err2 == nil {
				return nil
			}
		}
		return fmt.Errorf("memory: unmarshal failed: %w", err)
	}

	return nil
}

func BuildMemoryEntry(taskType, userInput, agentOutput string, keyDecisions []string, metadata map[string]string) *MemoryEntry {
	return BuildMemoryEntryWithEmbedder(context.Background(), taskType, userInput, agentOutput, keyDecisions, metadata, nil)
}

func BuildMemoryEntryWithEmbedder(ctx context.Context, taskType, userInput, agentOutput string,
	keyDecisions []string, metadata map[string]string, embedder Embedder) *MemoryEntry {
	if metadata == nil {
		metadata = make(map[string]string)
	}
	entry := &MemoryEntry{
		ID:           uuid.New().String(),
		CreatedAt:    time.Now(),
		TaskType:     taskType,
		UserInput:    userInput,
		AgentOutput:  agentOutput,
		KeyDecisions: keyDecisions,
		Metadata:     metadata,
	}
	if embedder != nil {
		if emb, err := embedder.Embed(ctx, userInput); err == nil {
			entry.Embedding = emb
			return entry
		}
	}
	entry.Embedding = mockEmbedding(userInput)
	return entry
}

func mockEmbedding(text string) []float32 {
	hash := sha256.Sum256([]byte(text))
	dim := 16
	vec := make([]float32, dim)
	for i := 0; i < dim; i++ {
		vec[i] = float32(hash[i%32]) / 255.0
	}
	return vec
}
