package dto

import "time"

type StudentQuery struct {
	Page      int
	PageSize  int
	StudentID string
	Name      string
	ClassID   uint
	Status    *uint8
}

// Student never exposes credentials or allows account identity to be edited.
type Student struct {
	ProfileRequest
	ID             uint       `json:"id"`
	StudentID      string     `json:"student_id"`
	Status         uint8      `json:"status"`
	AvatarURL      string     `json:"avatar_url"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastSeenAt     *time.Time `json:"last_seen_at"`
	LastSeenDevice string     `json:"last_seen_device"`
	AppVersion     string     `json:"app_version"`
}

type StudentListResponse struct {
	Items    []*StudentListItem `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// StudentListItem contains only fields displayed in the management table.
type StudentListItem struct {
	ID         uint       `json:"id"`
	StudentID  string     `json:"student_id"`
	Name       string     `json:"name"`
	ClassID    uint       `json:"class_id"`
	Gender     string     `json:"gender"`
	Phone      string     `json:"phone"`
	Status     uint8      `json:"status"`
	AppVersion string     `json:"app_version"`
	CreatedAt  time.Time  `json:"created_at"`
	LastSeenAt *time.Time `json:"last_seen_at"`
}

type StudentStatusRequest struct {
	Status *uint8 `json:"status"`
}
