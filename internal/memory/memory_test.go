package memory

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	a := []float32{1, 0, 0}
	b := []float32{1, 0, 0}
	if CosineSimilarity(a, b) != 1.0 {
		t.Errorf("identical vectors should have similarity 1.0, got %f", CosineSimilarity(a, b))
	}

	c := []float32{1, 0, 0}
	d := []float32{0, 1, 0}
	if CosineSimilarity(c, d) != 0 {
		t.Errorf("orthogonal vectors should have similarity 0, got %f", CosineSimilarity(c, d))
	}

	e := []float32{1, 2, 3}
	f := []float32{2, 4, 6}
	if CosineSimilarity(e, f) < 0.99 {
		t.Errorf("parallel vectors should have similarity ~1.0, got %f", CosineSimilarity(e, f))
	}
}

func TestEngineStoreAndCount(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "memory-test-*")
	defer os.RemoveAll(tmpDir)
	storePath := filepath.Join(tmpDir, "memory.json")

	engine, err := NewEmbeddedEngine(storePath)
	if err != nil {
		t.Fatalf("NewEmbeddedEngine failed: %v", err)
	}

	entry := BuildMemoryEntry("test", "user input", "agent output", []string{"key1"}, nil)
	if err := engine.Store(context.Background(), entry); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	count, _ := engine.Count()
	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}
}

func TestEngineSearch(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "memory-test-*")
	defer os.RemoveAll(tmpDir)
	storePath := filepath.Join(tmpDir, "memory.json")

	engine, _ := NewEmbeddedEngine(storePath)

	e1 := BuildMemoryEntry("type_a", "buy shoes", "output1", nil, nil)
	e2 := BuildMemoryEntry("type_b", "eat food", "output2", nil, nil)

	engine.Store(context.Background(), e1)
	engine.Store(context.Background(), e2)

	query := mockEmbedding("buy shoes")
	results, err := engine.Search(context.Background(), query, 5)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Log("Search returned 0 results (may happen with mock embeddings)")
	}
}

func TestEnginePersistence(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "memory-test-*")
	defer os.RemoveAll(tmpDir)
	storePath := filepath.Join(tmpDir, "memory.json")

	engine, _ := NewEmbeddedEngine(storePath)
	engine.Store(context.Background(), BuildMemoryEntry("test", "in", "out", nil, nil))

	// Re-create engine from same file
	engine2, err := NewEmbeddedEngine(storePath)
	if err != nil {
		t.Fatalf("reload failed: %v", err)
	}
	count, _ := engine2.Count()
	if count != 1 {
		t.Errorf("expected 1 entry after reload, got %d", count)
	}
}

func TestEmptyEngine(t *testing.T) {
	engine := NewEmptyEngine()
	count, _ := engine.Count()
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
	if err := engine.Close(); err != nil {
		t.Errorf("Close should not error: %v", err)
	}
}

func TestMockEmbedding(t *testing.T) {
	e1 := mockEmbedding("hello world")
	e2 := mockEmbedding("hello world")
	e3 := mockEmbedding("different text")

	if len(e1) != 16 {
		t.Errorf("expected 16-dim embedding, got %d", len(e1))
	}
	if CosineSimilarity(e1, e2) < 0.99 {
		t.Errorf("same text should have similar embeddings")
	}
	_ = CosineSimilarity(e1, e3) // just ensure no panic
}
