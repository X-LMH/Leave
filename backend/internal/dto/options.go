package dto

// ClassOption is a class item used by profile form selectors.
type ClassOption struct {
	ID        uint   `json:"id"`
	College   string `json:"college"`
	Major     string `json:"major"`
	ClassName string `json:"class_name"`
	IsEnabled bool   `json:"is_enabled"`
}

// ApartmentOption is an apartment item used by profile form selectors.
type ApartmentOption struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Gender    string `json:"gender"`
	IsEnabled bool   `json:"is_enabled"`
}

// LeaveTypeOption is a leave type item used by form selectors.
type LeaveTypeOption struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	IsEnabled bool   `json:"is_enabled"`
}
