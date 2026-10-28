package dto

// RateLimitResponse contains the seconds remaining before a request may be retried.
type RateLimitResponse struct {
	RetryAfter int64 `json:"retry_after"`
}
