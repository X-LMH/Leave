package dto

import "time"

// FeedbackCreateRequest 保存 JSON 或 multipart 提交中的反馈表单字段。
type FeedbackCreateRequest struct {
	Content     string `json:"content"`
	IsAnonymous bool   `json:"is_anonymous"`
}

type FeedbackResponse struct {
	ID          uint      `json:"id"`
	IsAnonymous bool      `json:"is_anonymous"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}

type FeedbackHandlingRequest struct {
	IsHandled *bool `json:"is_handled"`
}

type FeedbackAttachment struct {
	Index int    `json:"index"`
	URL   string `json:"url"`
}

type AdminFeedbackDetail struct {
	AdminFeedbackItem
	Attachments []FeedbackAttachment `json:"attachments"`
}

// AdminFeedbackItem keeps anonymous identities out of management responses.
type AdminFeedbackItem struct {
	ID          uint      `json:"id"`
	IsHandled   bool      `json:"is_handled"`
	StudentID   *string   `json:"student_id"`
	IsAnonymous bool      `json:"is_anonymous"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FeedbackListQuery struct {
	Page        int
	PageSize    int
	Keyword     string
	StudentID   string
	IsAnonymous *bool
	IsHandled   *bool
}

type FeedbackListResponse struct {
	Items    []*AdminFeedbackItem `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}
