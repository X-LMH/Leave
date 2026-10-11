package dto

import "time"

type ApartmentListQuery struct {
	Page      int
	PageSize  int
	Name      string
	Gender    string
	IsEnabled *bool
}

type ApartmentRequest struct {
	Name      string  `json:"name"`
	Gender    string  `json:"gender"`
	SortOrder *uint32 `json:"sort_order"`
	IsEnabled *bool   `json:"is_enabled"`
}

type ApartmentListItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Gender    string    `json:"gender"`
	SortOrder uint      `json:"sort_order"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ApartmentListResponse struct {
	Items    []*ApartmentListItem `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}
