package publish

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/wjames2000/opc-macs/internal/platform"
)

// PublishStatus tracks the status of a publish job.
type PublishStatus string

const (
	StatusQueued     PublishStatus = "queued"
	StatusPublishing PublishStatus = "publishing"
	StatusSuccess    PublishStatus = "success"
	StatusPartial    PublishStatus = "partial"
	StatusFailed     PublishStatus = "failed"
	StatusCancelled  PublishStatus = "cancelled"
)

// PublishJob represents a single publish operation to one platform account.
type PublishJob struct {
	ID          string
	AccountID   string
	Platform    string
	Request     *platform.PublishRequest
	Status      PublishStatus
	Result      *platform.PublishResponse
	Error       error
	RetryCount  int
	MaxRetries  int
	CreatedAt   time.Time
	ScheduledAt *time.Time
	CompletedAt *time.Time
}

// Engine manages publishing content to multiple platform accounts.
// It handles queue management, retry logic, content adaptation, and scheduling.
type Engine struct {
	mu      sync.RWMutex
	clients map[string]platform.PlatformClient // platform name -> client
	jobs    map[string]*PublishJob
	queue   []*PublishJob
	workers int
	workCh  chan *PublishJob
	stopCh  chan struct{}
	wg      sync.WaitGroup

	maxRetries int
	baseDelay  time.Duration
}

// NewEngine creates a new publish engine with the given number of workers.
func NewEngine(workers int) *Engine {
	if workers < 1 {
		workers = 3
	}
	return &Engine{
		clients:    make(map[string]platform.PlatformClient),
		jobs:       make(map[string]*PublishJob),
		workers:    workers,
		workCh:     make(chan *PublishJob, 100),
		stopCh:     make(chan struct{}),
		maxRetries: 3,
		baseDelay:  5 * time.Second,
	}
}

// RegisterClient registers a platform client for publishing.
func (e *Engine) RegisterClient(client platform.PlatformClient) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.clients[client.Name()] = client
	log.Printf("[Publish] Registered client: %s", client.Name())
}

// Publish submits a single publish request to one platform.
func (e *Engine) Publish(ctx context.Context, accountID string, req *platform.PublishRequest) (*PublishJob, error) {
	job := &PublishJob{
		ID:          fmt.Sprintf("pub_%d", time.Now().UnixNano()),
		AccountID:   accountID,
		Request:     req,
		Status:      StatusQueued,
		MaxRetries:  e.maxRetries,
		CreatedAt:   time.Now(),
		ScheduledAt: req.ScheduleAt,
	}

	e.mu.Lock()
	client, ok := e.clients[req.PlatformSpecific["platform"].(string)]
	if !ok {
		e.mu.Unlock()
		return nil, fmt.Errorf("no client registered for platform")
	}
	e.jobs[job.ID] = job
	e.mu.Unlock()

	if req.ScheduleAt != nil && req.ScheduleAt.After(time.Now()) {
		log.Printf("[Publish] Scheduled job %s for %s", job.ID, req.ScheduleAt.Format(time.RFC3339))
		go e.scheduleJob(job, client)
	} else {
		e.workCh <- job
	}

	return job, nil
}

// PublishMulti dispatches a single content to multiple platforms.
func (e *Engine) PublishMulti(ctx context.Context, accounts []string, req *platform.PublishRequest) ([]*PublishJob, error) {
	var jobs []*PublishJob
	for _, accountID := range accounts {
		job, err := e.Publish(ctx, accountID, req)
		if err != nil {
			log.Printf("[Publish] Failed to submit job for account %s: %v", accountID, err)
			continue
		}
		jobs = append(jobs, job)
	}
	if len(jobs) == 0 {
		return nil, fmt.Errorf("no jobs submitted")
	}
	return jobs, nil
}

// GetJob returns the status of a publish job.
func (e *Engine) GetJob(jobID string) (*PublishJob, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	job, ok := e.jobs[jobID]
	return job, ok
}

// ListJobs returns all jobs with optional status filter.
func (e *Engine) ListJobs(status PublishStatus) []*PublishJob {
	e.mu.RLock()
	defer e.mu.RUnlock()
	var result []*PublishJob
	for _, job := range e.jobs {
		if status == "" || job.Status == status {
			result = append(result, job)
		}
	}
	return result
}

// CancelJob cancels a queued or in-progress job.
func (e *Engine) CancelJob(jobID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	job, ok := e.jobs[jobID]
	if !ok {
		return fmt.Errorf("job %s not found", jobID)
	}
	if job.Status == StatusQueued {
		job.Status = StatusCancelled
	}
	return nil
}

// Start launches the worker pool.
func (e *Engine) Start() {
	for i := 0; i < e.workers; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}
	log.Printf("[Publish] Engine started with %d workers", e.workers)
}

// Stop gracefully shuts down the engine.
func (e *Engine) Stop() {
	close(e.stopCh)
	e.wg.Wait()
	log.Println("[Publish] Engine stopped")
}

func (e *Engine) worker(id int) {
	defer e.wg.Done()
	for {
		select {
		case job := <-e.workCh:
			e.processJob(job)
		case <-e.stopCh:
			return
		}
	}
}

func (e *Engine) processJob(job *PublishJob) {
	e.mu.Lock()
	platformName := job.Request.PlatformSpecific["platform"].(string)
	client, ok := e.clients[platformName]
	job.Status = StatusPublishing
	e.mu.Unlock()

	if !ok {
		job.Status = StatusFailed
		job.Error = fmt.Errorf("no client registered for platform: %s", platformName)
		return
	}

	result, err := client.Publish(context.Background(), job.Request)
	if err != nil {
		if job.RetryCount < job.MaxRetries {
			job.RetryCount++
			delay := e.baseDelay * time.Duration(1<<uint(job.RetryCount-1))
			log.Printf("[Publish] Job %s failed (retry %d/%d), retrying in %s: %v",
				job.ID, job.RetryCount, job.MaxRetries, delay, err)
			time.AfterFunc(delay, func() {
				e.workCh <- job
			})
			return
		}
		job.Status = StatusFailed
		job.Error = err
		job.CompletedAt = timePtr(time.Now())
		log.Printf("[Publish] Job %s failed after %d retries: %v", job.ID, job.MaxRetries, err)
		return
	}

	if result != nil {
		job.Status = PublishStatus(result.Status)
		job.Result = result
	} else {
		job.Status = StatusSuccess
	}
	job.CompletedAt = timePtr(time.Now())
	log.Printf("[Publish] Job %s completed: status=%s, post_id=%s",
		job.ID, job.Status, job.Result)
}

func (e *Engine) scheduleJob(job *PublishJob, client platform.PlatformClient) {
	delay := time.Until(*job.ScheduledAt)
	if delay > 0 {
		time.Sleep(delay)
	}
	e.workCh <- job
}

func timePtr(t time.Time) *time.Time {
	return &t
}
