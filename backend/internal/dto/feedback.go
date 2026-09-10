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
