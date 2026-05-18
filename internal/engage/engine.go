package engage

import (
	"context"
	"sync"
)

type MetricsCollector interface {
	Collect(ctx context.Context, platform, postID string) (*EngagementMetrics, error)
}

type CommentFetcher interface {
	FetchComments(ctx context.Context, platform, postID string) ([]*Comment, error)
	PostReply(ctx context.Context, req *ReplyRequest) error
}

type Engine struct {
	collector MetricsCollector
	fetcher   CommentFetcher
	mu        sync.RWMutex
	metrics   map[string]*EngagementMetrics
}

func NewEngine(collector MetricsCollector, fetcher CommentFetcher) *Engine {
	return &Engine{
		collector: collector,
		fetcher:   fetcher,
		metrics:   make(map[string]*EngagementMetrics),
	}
}

func (e *Engine) FetchComments(ctx context.Context, platform, postID string) ([]*Comment, error) {
	return e.fetcher.FetchComments(ctx, platform, postID)
}

func (e *Engine) Reply(ctx context.Context, req *ReplyRequest) error {
	return e.fetcher.PostReply(ctx, req)
}

func (e *Engine) TrackEngagement(ctx context.Context, platform, postID string) (*EngagementMetrics, error) {
	m, err := e.collector.Collect(ctx, platform, postID)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	key := platform + ":" + postID
	if existing, ok := e.metrics[key]; ok {
		m.EngagementRate = computeRate(m, existing)
	} else {
		total := m.Likes + m.Comments + m.Shares + m.Saves
		if m.Views > 0 {
			m.EngagementRate = float64(total) / float64(m.Views) * 100
		}
	}
	e.metrics[key] = m
	e.mu.Unlock()

	return m, nil
}

func computeRate(current, previous *EngagementMetrics) float64 {
	newTotal := (current.Likes + current.Comments + current.Shares + current.Saves) -
		(previous.Likes + previous.Comments + previous.Shares + previous.Saves)
	newViews := current.Views - previous.Views
	if newViews <= 0 {
		return current.EngagementRate
	}
	return float64(newTotal) / float64(newViews) * 100
}

func autoReply(ctx context.Context, engine *Engine, comment *Comment, template string) {
	reply := map[string]string{
		"like":     "感谢您的喜欢！",
		"question": "感谢提问！更多信息请关注我们的主页。",
		"praise":   "谢谢支持！我们会继续努力。",
		"default":  "感谢您的评论！",
	}

	content, ok := reply[comment.Content]
	if !ok {
		content = reply["default"]
	}

	if template != "" {
		content = template
	}

	engine.Reply(ctx, &ReplyRequest{
		Platform:  comment.Platform,
		CommentID: comment.ID,
		Content:   content,
	})
}
