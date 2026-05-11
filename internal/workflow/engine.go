package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wjames2000/opc-macs/internal/runtime"
)

type Engine struct {
	loader *runtime.Loader
	client runtime.ModelClient
}

func NewEngine(loader *runtime.Loader, client runtime.ModelClient) *Engine {
	return &Engine{loader: loader, client: client}
}

func (e *Engine) Run(ctx context.Context, wf *Workflow, input string) (*Session, error) {
	session := &Session{
		ID:        fmt.Sprintf("wf-%d", time.Now().UnixNano()),
		Workflow:  wf,
		Status:    "running",
		StartTime: time.Now(),
		Context:   make(map[string]interface{}),
	}

	// Set trigger input
	session.SetVar("trigger", map[string]interface{}{
		"input": input,
	})
	session.SetVar("workflow", map[string]interface{}{
		"name":       wf.Name,
		"started_at": time.Now().Format(time.RFC3339),
	})

	for _, step := range wf.Steps {
		select {
		case <-ctx.Done():
			session.Status = "cancelled"
			return session, ctx.Err()
		default:
		}

		stepLog := StepLog{
			StepID:    step.ID,
			Status:    "running",
			StartTime: time.Now(),
		}

		// Check condition
		if step.Condition != "" {
			pass, _ := evalCondition(step.Condition, session.Context)
			if !pass {
				stepLog.Status = "skipped"
				stepLog.Duration = time.Since(stepLog.StartTime)
				session.AddLog(stepLog)
				continue
			}
		}

		if len(step.Parallel) > 0 {
			// Parallel execution
			stepLog.Agent = "parallel"
			result, err := e.executeParallel(ctx, step, session)
			stepLog.Duration = time.Since(stepLog.StartTime)
			if err != nil {
				stepLog.Status = "failed"
				stepLog.Error = err.Error()
				session.Status = "failed"
				session.AddLog(stepLog)
				session.EndTime = time.Now()
				return session, fmt.Errorf("workflow %s step %s failed: %w", wf.Name, step.ID, err)
			}
			stepLog.Status = "completed"
			if result != nil {
				stepLog.Output = fmt.Sprintf("%+v", result)
			}
		} else {
			// Serial execution
			stepLog.Agent = step.Agent
			result, err := e.executeSerial(ctx, step, session)
			stepLog.Duration = time.Since(stepLog.StartTime)
			if err != nil {
				stepLog.Status = "failed"
				stepLog.Error = err.Error()
				session.Status = "failed"
				session.AddLog(stepLog)
				session.EndTime = time.Now()
				return session, fmt.Errorf("workflow %s step %s failed: %w", wf.Name, step.ID, err)
			}
			stepLog.Status = "completed"
			if result != nil {
				stepLog.Output = fmt.Sprintf("%+v", result)
			}
		}

		session.AddLog(stepLog)
	}

	session.Status = "completed"
	session.EndTime = time.Now()
	return session, nil
}

func (e *Engine) executeSerial(ctx context.Context, step *Step, session *Session) (interface{}, error) {
	input := ResolveVariables(step.Input, session.Context)
	plugin, ok := e.loader.Get(step.Agent)
	if !ok {
		return nil, fmt.Errorf("agent %s not found", step.Agent)
	}

	opts := map[string]interface{}{
		"model_client": e.client,
	}

	var result *runtime.ExecutionResult
	var err error

	if step.Retry != nil && step.Retry.MaxAttempts > 0 {
		for i := 0; i <= step.Retry.MaxAttempts; i++ {
			result, err = plugin.Execute(ctx, input, opts)
			if err == nil {
				break
			}
			if i < step.Retry.MaxAttempts && step.Retry.DelayMs > 0 {
				time.Sleep(time.Duration(step.Retry.DelayMs) * time.Millisecond)
			}
		}
	} else {
		result, err = plugin.Execute(ctx, input, opts)
	}

	if err != nil {
		return nil, err
	}

	if step.Output != "" {
		session.SetVar(step.Output, result.Data)
	}
	return result.Data, nil
}

func (e *Engine) executeParallel(ctx context.Context, step *Step, session *Session) (interface{}, error) {
	var wg sync.WaitGroup
	errCh := make(chan error, len(step.Parallel))
	results := make(map[string]interface{})
	var mu sync.Mutex

	for _, ps := range step.Parallel {
		wg.Add(1)
		go func(p *ParallelStep) {
			defer wg.Done()
			input := ResolveVariables(p.Input, session.Context)
			plugin, ok := e.loader.Get(p.Agent)
			if !ok {
				errCh <- fmt.Errorf("agent %s not found", p.Agent)
				return
			}
			result, err := plugin.Execute(ctx, input, map[string]interface{}{
				"model_client": e.client,
			})
			if err != nil {
				errCh <- fmt.Errorf("%s: %w", p.ID, err)
				return
			}
			mu.Lock()
			results[p.ID] = result.Data
			if p.Output != "" {
				session.SetVar(p.Output, result.Data)
			}
			mu.Unlock()
		}(ps)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}
