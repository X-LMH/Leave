package dto

// ProfileRequest is the JSON body accepted by the profile create or update endpoint.
type ProfileRequest struct {
	Name            string `json:"name"`
	Phone           string `json:"phone"`
	Gender          string `json:"gender"`
	ParentName      string `json:"parent_name"`
	ParentPhone     string `json:"parent_phone"`
	ClassID         uint   `json:"class_id"`
	ApartmentID     uint   `json:"apartment_id"`
	DormitoryNumber string `json:"dormitory_number"`
	TeacherName     string `json:"teacher_name"`
}

// ProfileClassInfo contains display data for the student's class.
type ProfileClassInfo struct {
	College   string `json:"college"`
	Major     string `json:"major"`
	ClassName string `json:"class_name"`
}

// ProfileApartmentInfo contains display data for the student's residence.
type ProfileApartmentInfo struct {
	ApartmentName   string `json:"apartment_name"`
	DormitoryNumber string `json:"dormitory_number"`
}

// ProfileResponse is the profile data returned by the profile endpoint.
type ProfileResponse struct {
	StudentID     string               `json:"student_id"`
	Name          string               `json:"name"`
	Phone         string               `json:"phone"`
	Gender        string               `json:"gender"`
	ParentName    string               `json:"parent_name"`
	ParentPhone   string               `json:"parent_phone"`
	TeacherName   string               `json:"teacher_name"`
	ClassInfo     ProfileClassInfo     `json:"class_info"`
	ApartmentInfo ProfileApartmentInfo `json:"apartment_info"`
}
