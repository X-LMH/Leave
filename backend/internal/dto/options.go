package dto

import "time"

// ClassOption is a class item used by profile form selectors.
type ClassOption struct {
	ID        uint   `json:"id"`
	College   string `json:"college"`
	Major     string `json:"major"`
	ClassName string `json:"class_name"`
}

type AdminClassListQuery struct {
	ClassName string
	Page      int
	PageSize  int
	College   string
	Major     string
	IsEnabled *bool
}

type AdminClassListItem struct {
	ID        uint      `json:"id"`
	College   string    `json:"college"`
	Major     string    `json:"major"`
	ClassName string    `json:"class_name"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AdminClassListResponse struct {
	Items    []*AdminClassListItem `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// ApartmentOption is an enabled apartment item used by profile form selectors.
type ApartmentOption struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

// LeaveTypeOption is an enabled leave type used by the leave application form.
type LeaveTypeOption struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}
