package saas

import (
	"encoding/json"
	"net/http"
	"time"
)

type PublishHandler struct {
}

func NewPublishHandler() *PublishHandler {
	return &PublishHandler{}
}

type publishRequest struct {
	Platform   string   `json:"platform"`
	Title      string   `json:"title"`
	Body       string   `json:"body"`
	MediaURLs  []string `json:"media_urls,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	ScheduleAt string   `json:"schedule_at,omitempty"`
}

type publishResponse struct {
	TaskID      string `json:"task_id"`
	Platform    string `json:"platform"`
	Status      string `json:"status"`
	PublishedAt string `json:"published_at,omitempty"`
}

func (h *PublishHandler) HandlePublish(w http.ResponseWriter, r *http.Request) {
	var req publishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Platform == "" {
		writeError(w, http.StatusBadRequest, "platform is required")
		return
	}
	if req.Title == "" && req.Body == "" {
		writeError(w, http.StatusBadRequest, "title or body is required")
		return
	}

	resp := publishResponse{
		TaskID:      GenerateID(),
		Platform:    req.Platform,
		Status:      "queued",
		PublishedAt: "",
	}
	if req.ScheduleAt != "" {
		resp.Status = "scheduled"
	}

	writeJSON(w, http.StatusAccepted, resp)
}

type scheduleListResponse struct {
	Schedules []scheduleItem `json:"schedules"`
}

type scheduleItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Platform  string `json:"platform"`
	Status    string `json:"status"`
	Scheduled string `json:"scheduled_at"`
	CreatedAt string `json:"created_at"`
}

func (h *PublishHandler) HandleListSchedules(w http.ResponseWriter, r *http.Request) {
	resp := scheduleListResponse{
		Schedules: []scheduleItem{
			{
				ID:        GenerateID(),
				Title:     "示例: 小红书笔记",
				Platform:  "xiaohongshu",
				Status:    "scheduled",
				Scheduled: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				CreatedAt: time.Now().Format(time.RFC3339),
			},
		},
	}
	writeJSON(w, http.StatusOK, resp)
}
