package publish

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LogEntry represents a single publish log record.
type LogEntry struct {
	ID             string
	JobID          string
	AccountID      string
	Platform       string
	Content        string
	Status         PublishStatus
	Error          string
	PostURL        string
	PlatformPostID string
	RetryCount     int
	CreatedAt      time.Time
	CompletedAt    *time.Time
}

// LogStore defines the interface for publish log persistence.
type LogStore interface {
	Save(ctx context.Context, entry *LogEntry) error
	GetByJobID(ctx context.Context, jobID string) (*LogEntry, error)
	GetByPlatform(ctx context.Context, platform string, limit, offset int) ([]*LogEntry, error)
	GetByAccount(ctx context.Context, accountID string, limit, offset int) ([]*LogEntry, error)
	List(ctx context.Context, limit, offset int) ([]*LogEntry, error)
}

// MemoryLogStore implements LogStore with in-memory storage.
// Used for development/testing. Production should use a database-backed store.
type MemoryLogStore struct {
	mu   sync.RWMutex
	logs map[string]*LogEntry
}

func NewMemoryLogStore() *MemoryLogStore {
	return &MemoryLogStore{
		logs: make(map[string]*LogEntry),
	}
}

func (s *MemoryLogStore) Save(ctx context.Context, entry *LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("log_%d", time.Now().UnixNano())
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	s.logs[entry.ID] = entry
	return nil
}

func (s *MemoryLogStore) GetByJobID(ctx context.Context, jobID string) (*LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, entry := range s.logs {
		if entry.JobID == jobID {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("log not found for job: %s", jobID)
}

func (s *MemoryLogStore) GetByPlatform(ctx context.Context, platform string, limit, offset int) ([]*LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*LogEntry
	for _, entry := range s.logs {
		if entry.Platform == platform {
			result = append(result, entry)
		}
	}
	return applyLimitOffset(result, limit, offset), nil
}

func (s *MemoryLogStore) GetByAccount(ctx context.Context, accountID string, limit, offset int) ([]*LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*LogEntry
	for _, entry := range s.logs {
		if entry.AccountID == accountID {
			result = append(result, entry)
		}
	}
	return applyLimitOffset(result, limit, offset), nil
}

func (s *MemoryLogStore) List(ctx context.Context, limit, offset int) ([]*LogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*LogEntry
	for _, entry := range s.logs {
		result = append(result, entry)
	}
	return applyLimitOffset(result, limit, offset), nil
}

func applyLimitOffset(entries []*LogEntry, limit, offset int) []*LogEntry {
	if offset >= len(entries) {
		return nil
	}
	end := offset + limit
	if end > len(entries) || limit == 0 {
		end = len(entries)
	}
	return entries[offset:end]
}
