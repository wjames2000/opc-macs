package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	pgvector "github.com/pgvector/pgvector-go"
)

type PGVectorStore struct {
	db *sql.DB
}

func NewPGVectorStore(connStr string) (*PGVectorStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("pgvector: open failed: %w", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pgvector: ping failed: %w", err)
	}

	// Auto-create extension and table
	if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS vector`); err != nil {
		return nil, fmt.Errorf("pgvector: create extension: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS memory_entries (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			task_type   TEXT NOT NULL,
			user_input  TEXT NOT NULL,
			agent_output TEXT NOT NULL,
			key_decisions TEXT[],
			embedding   VECTOR(768),
			metadata    JSONB DEFAULT '{}'::jsonb
		)
	`); err != nil {
		return nil, fmt.Errorf("pgvector: create table: %w", err)
	}

	// Create index if not exists (IVFFlat with cosine similarity)
	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_memory_entries_embedding
			ON memory_entries
			USING ivfflat (embedding vector_cosine_ops)
			WITH (lists = 100)
	`); err != nil {
		// Index creation may fail on small datasets, ignore
		_ = err
	}

	return &PGVectorStore{db: db}, nil
}

func (p *PGVectorStore) Store(ctx context.Context, entry *MemoryEntry) error {
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("%x", time.Now().UnixNano())
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	metaJSON, _ := json.Marshal(entry.Metadata)
	if metaJSON == nil {
		metaJSON = []byte("{}")
	}

	vec := pgvector.NewVector(entry.Embedding)
	keyDecisions := make(pq.StringArray, len(entry.KeyDecisions))
	for i, kd := range entry.KeyDecisions {
		keyDecisions[i] = kd
	}

	_, err := p.db.ExecContext(ctx, `
		INSERT INTO memory_entries (id, created_at, task_type, user_input, agent_output, key_decisions, embedding, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO NOTHING
	`, entry.ID, entry.CreatedAt, entry.TaskType, entry.UserInput, entry.AgentOutput,
		keyDecisions, vec, metaJSON)

	return err
}

func (p *PGVectorStore) Search(ctx context.Context, query []float32, topK int) ([]*MemoryEntry, error) {
	if topK <= 0 {
		topK = 5
	}

	vec := pgvector.NewVector(query)
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, created_at, task_type, user_input, agent_output, key_decisions, metadata,
			   1 - (embedding <=> $1) AS similarity
		FROM memory_entries
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $2
	`, vec, topK)
	if err != nil {
		return nil, fmt.Errorf("pgvector: search failed: %w", err)
	}
	defer rows.Close()

	var results []*MemoryEntry
	for rows.Next() {
		var (
			entry       MemoryEntry
			metaJSON    []byte
			keyDecisions pq.StringArray
			similarity  float64
		)
		if err := rows.Scan(&entry.ID, &entry.CreatedAt, &entry.TaskType,
			&entry.UserInput, &entry.AgentOutput, &keyDecisions,
			&metaJSON, &similarity); err != nil {
			return nil, fmt.Errorf("pgvector: scan failed: %w", err)
		}
		entry.KeyDecisions = []string(keyDecisions)
		if len(metaJSON) > 0 {
			json.Unmarshal(metaJSON, &entry.Metadata)
		}
		if entry.Metadata == nil {
			entry.Metadata = make(map[string]string)
		}
		entry.Metadata["similarity"] = fmt.Sprintf("%.4f", similarity)
		results = append(results, &entry)
	}

	return results, nil
}

func (p *PGVectorStore) Recall(ctx context.Context, text string, topK int) ([]*MemoryEntry, error) {
	embedding := mockEmbedding(text)
	return p.Search(ctx, embedding, topK)
}

func (p *PGVectorStore) Count() (int, error) {
	var count int
	if err := p.db.QueryRow(`SELECT COUNT(*) FROM memory_entries`).Scan(&count); err != nil {
		return 0, fmt.Errorf("pgvector: count failed: %w", err)
	}
	return count, nil
}

func (p *PGVectorStore) Close() error {
	return p.db.Close()
}
