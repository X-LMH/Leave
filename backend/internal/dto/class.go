package dto

import "time"

type ClassListQuery struct {
	ClassName string
	Page      int
	PageSize  int
	College   string
	Major     string
	IsEnabled *bool
}

type ClassListItem struct {
	ID        uint      `json:"id"`
	College   string    `json:"college"`
	Major     string    `json:"major"`
	ClassName string    `json:"class_name"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ClassListResponse struct {
	Items    []*ClassListItem `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type ClassRequest struct {
	College   string `json:"college"`
	Major     string `json:"major"`
	ClassName string `json:"class_name"`
	IsEnabled *bool  `json:"is_enabled"`
}
