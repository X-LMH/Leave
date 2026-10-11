package dto

import "time"

// FeedbackCreateRequest is the JSON body accepted when submitting feedback.
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

// AdminFeedbackItem keeps anonymous identities out of management responses.
type AdminFeedbackItem struct {
	ID          uint      `json:"id"`
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
}

type FeedbackListResponse struct {
	Items    []*AdminFeedbackItem `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}
