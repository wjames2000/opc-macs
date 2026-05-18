package media

import (
	"context"
	"fmt"
	"time"
)

// VideoPoller polls a video generation service for job completion.
type VideoPoller struct {
	interval time.Duration
	timeout  time.Duration
}

// NewVideoPoller creates a VideoPoller with default settings (5s interval, 5min timeout).
func NewVideoPoller() *VideoPoller {
	return &VideoPoller{
		interval: 5 * time.Second,
		timeout:  5 * time.Minute,
	}
}

// NewVideoPollerWithConfig creates a VideoPoller with custom settings.
func NewVideoPollerWithConfig(interval, timeout time.Duration) *VideoPoller {
	return &VideoPoller{
		interval: interval,
		timeout:  timeout,
	}
}

// Poll repeatedly checks the status of a video generation job until it completes or fails.
// client must implement CheckStatus to query the generation service.
func (p *VideoPoller) Poll(ctx context.Context, jobID string, client StatusChecker) (*GenerationResult, error) {
	if jobID == "" {
		return nil, fmt.Errorf("video_poller: empty jobID")
	}
	if client == nil {
		return nil, fmt.Errorf("video_poller: nil client")
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return nil, fmt.Errorf("video_poller: timeout after %v waiting for job %s", p.timeout, jobID)
		case <-ticker.C:
			result, err := client.CheckStatus(timeoutCtx, jobID)
			if err != nil {
				return nil, fmt.Errorf("video_poller: check status failed for %s: %w", jobID, err)
			}

			switch result.Status {
			case StatusCompleted:
				return result, nil
			case StatusFailed:
				return result, fmt.Errorf("video_poller: job %s failed: %s", jobID, result.Error)
			case StatusPending, StatusProcessing:
				// continue polling
			}
		}
	}
}
