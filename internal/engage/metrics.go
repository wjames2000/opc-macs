package engage

import "time"

type EngagementMetrics struct {
	PostID         string    `json:"post_id"`
	Platform       Platform  `json:"platform"`
	Likes          int64     `json:"likes"`
	Comments       int64     `json:"comments"`
	Shares         int64     `json:"shares"`
	Saves          int64     `json:"saves"`
	Views          int64     `json:"views"`
	Followers      int64     `json:"followers"`
	EngagementRate float64   `json:"engagement_rate"`
	CollectedAt    time.Time `json:"collected_at"`
}

type MetricsSnapshot struct {
	ID             string            `json:"id"`
	PostID         string            `json:"post_id"`
	Platform       Platform          `json:"platform"`
	Metrics        EngagementMetrics `json:"metrics"`
	EngagementRate float64           `json:"engagement_rate"`
	Period         string            `json:"period"`
	CollectedAt    time.Time         `json:"collected_at"`
}

func computeEngagementRate(m *EngagementMetrics) float64 {
	if m.Views == 0 {
		return 0
	}
	total := float64(m.Likes + m.Comments + m.Shares + m.Saves)
	return (total / float64(m.Views)) * 100
}

func computeEngagementRateWeighted(m *EngagementMetrics) float64 {
	if m.Views == 0 {
		return 0
	}
	weighted := float64(m.Likes)*1 + float64(m.Comments)*3 + float64(m.Shares)*5 + float64(m.Saves)*4
	return (weighted / float64(m.Views)) * 100
}
