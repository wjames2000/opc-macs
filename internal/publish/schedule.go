package publish

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ScheduleEntry represents a scheduled publish job.
type ScheduleEntry struct {
	ID          string
	JobID       string
	AccountID   string
	Platform    string
	Content     string // serialized content reference
	ScheduledAt time.Time
	CreatedAt   time.Time
	Status      string // pending, published, cancelled, missed
}

// ScheduleManager handles scheduled (future-dated) publish operations.
type ScheduleManager struct {
	mu      sync.RWMutex
	engine  *Engine
	entries map[string]*ScheduleEntry
	stopCh  chan struct{}
}

// NewScheduleManager creates a new schedule manager.
func NewScheduleManager(engine *Engine) *ScheduleManager {
	return &ScheduleManager{
		engine:  engine,
		entries: make(map[string]*ScheduleEntry),
		stopCh:  make(chan struct{}),
	}
}

// Schedule adds a new scheduled publish entry.
func (sm *ScheduleManager) Schedule(ctx context.Context, entry *ScheduleEntry) error {
	if entry.ScheduledAt.Before(time.Now()) {
		return fmt.Errorf("schedule time must be in the future")
	}

	sm.mu.Lock()
	entry.ID = fmt.Sprintf("sched_%d", time.Now().UnixNano())
	entry.CreatedAt = time.Now()
	entry.Status = "pending"
	sm.entries[entry.ID] = entry
	sm.mu.Unlock()

	log.Printf("[Schedule] Entry %s scheduled for %s", entry.ID, entry.ScheduledAt.Format(time.RFC3339))
	return nil
}

// Cancel removes a scheduled entry.
func (sm *ScheduleManager) Cancel(entryID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	entry, ok := sm.entries[entryID]
	if !ok {
		return fmt.Errorf("schedule entry %s not found", entryID)
	}
	if entry.Status == "published" || entry.Status == "cancelled" {
		return fmt.Errorf("cannot cancel entry with status: %s", entry.Status)
	}
	entry.Status = "cancelled"
	return nil
}

// ListPending returns all pending scheduled entries.
func (sm *ScheduleManager) ListPending() []*ScheduleEntry {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	var result []*ScheduleEntry
	now := time.Now()
	for _, entry := range sm.entries {
		if entry.Status == "pending" {
			if entry.ScheduledAt.Before(now) {
				entry.Status = "missed"
				continue
			}
			result = append(result, entry)
		}
	}
	return result
}

// Start launches the scheduler loop that checks for due entries.
func (sm *ScheduleManager) Start() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sm.processDueEntries()
			case <-sm.stopCh:
				return
			}
		}
	}()
	log.Println("[Schedule] Manager started")
}

// Stop stops the scheduler loop.
func (sm *ScheduleManager) Stop() {
	close(sm.stopCh)
}

func (sm *ScheduleManager) processDueEntries() {
	sm.mu.RLock()
	var dueEntries []*ScheduleEntry
	now := time.Now()
	for _, entry := range sm.entries {
		if entry.Status == "pending" && entry.ScheduledAt.Before(now) {
			dueEntries = append(dueEntries, entry)
		}
	}
	sm.mu.RUnlock()

	for _, entry := range dueEntries {
		log.Printf("[Schedule] Processing due entry %s (scheduled at %s)",
			entry.ID, entry.ScheduledAt.Format(time.RFC3339))
		// Submit to publish engine
		// The actual submit would use the stored content reference
		// For now, mark as published after submission
		sm.mu.Lock()
		entry.Status = "published"
		sm.mu.Unlock()
	}
}
