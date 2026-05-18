package engage

import "time"

type Platform string

const (
	PlatformDouyin      Platform = "douyin"
	PlatformXiaohongshu Platform = "xiaohongshu"
	PlatformBilibili    Platform = "bilibili"
	PlatformKuaishou    Platform = "kuaishou"
	PlatformWechat      Platform = "wechat"
	PlatformYoutube     Platform = "youtube"
	PlatformTiktok      Platform = "tiktok"
)

type InteractionType string

const (
	InteractionLike    InteractionType = "like"
	InteractionComment InteractionType = "comment"
	InteractionShare   InteractionType = "share"
	InteractionSave    InteractionType = "save"
	InteractionFollow  InteractionType = "follow"
)

type Comment struct {
	ID          string    `json:"id"`
	Platform    Platform  `json:"platform"`
	PostID      string    `json:"post_id"`
	CommentID   string    `json:"comment_id"`
	AuthorID    string    `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	Content     string    `json:"content"`
	ParentID    string    `json:"parent_id,omitempty"`
	LikeCount   int64     `json:"like_count"`
	ReplyCount  int64     `json:"reply_count"`
	CreatedAt   time.Time `json:"created_at"`
	CollectedAt time.Time `json:"collected_at"`
}

type ReplyRequest struct {
	Platform  Platform `json:"platform"`
	PostID    string   `json:"post_id"`
	CommentID string   `json:"comment_id"`
	Content   string   `json:"content"`
}

type ReplyResponse struct {
	ReplyID   string    `json:"reply_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentFilter struct {
	Platforms []Platform `json:"platforms,omitempty"`
	Statuses  []string   `json:"statuses,omitempty"`
	Keywords  []string   `json:"keywords,omitempty"`
	Since     *time.Time `json:"since,omitempty"`
	Limit     int        `json:"limit,omitempty"`
}

type AutoReplyRule struct {
	ID        string   `json:"id"`
	Platforms []string `json:"platforms"`
	Keywords  []string `json:"keywords"`
	ReplyText string   `json:"reply_text"`
	Enabled   bool     `json:"enabled"`
	Priority  int      `json:"priority"`
}
