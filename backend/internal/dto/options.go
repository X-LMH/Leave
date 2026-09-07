package dto

// ClassOption is a class item used by profile form selectors.
type ClassOption struct {
	ID        uint   `json:"id"`
	College   string `json:"college"`
	Major     string `json:"major"`
	ClassName string `json:"class_name"`
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
