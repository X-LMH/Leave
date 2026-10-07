package dto

import "time"

type LeaveTypeListQuery struct {
	Page      int
	PageSize  int
	Name      string
	IsEnabled *bool
}
type LeaveTypeRequest struct {
	Name      string  `json:"name"`
	SortOrder *uint64 `json:"sort_order"`
	IsEnabled *bool   `json:"is_enabled"`
}

type LeaveTypeItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	SortOrder uint      `json:"sort_order"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type LeaveTypeListResponse struct {
	Items    []*LeaveTypeItem `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}
