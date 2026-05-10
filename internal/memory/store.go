package memory

import "context"

type MemoryStore interface {
	Store(ctx context.Context, entry *MemoryEntry) error
	Search(ctx context.Context, query []float32, topK int) ([]*MemoryEntry, error)
	Recall(ctx context.Context, text string, topK int) ([]*MemoryEntry, error)
	Count() (int, error)
	Close() error
}
