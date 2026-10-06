package dto

import "time"

type AdminLeaveTypeListQuery struct {
	Page      int
	PageSize  int
	Name      string
	IsEnabled *bool
}
type AdminLeaveTypeRequest struct {
	Name      string  `json:"name"`
	SortOrder *uint64 `json:"sort_order"`
	IsEnabled *bool   `json:"is_enabled"`
}
type AdminClassRequest struct {
	College   string `json:"college"`
	Major     string `json:"major"`
	ClassName string `json:"class_name"`
	IsEnabled *bool  `json:"is_enabled"`
}
type AdminLeaveTypeItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	SortOrder uint      `json:"sort_order"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type AdminLeaveTypeListResponse struct {
	Items    []*AdminLeaveTypeItem `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}
