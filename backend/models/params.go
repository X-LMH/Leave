package models

import "time"

type ParamRegister struct {
	StudentID  string `json:"student_id"`
	Password   string `json:"password"`
	RePassword string `json:"re_password"`
}

type ParamLogin struct {
	StudentID string `json:"student_id"`
	Password  string `json:"password"`
}

type ParamProfile struct {
	Name        string `json:"name" gorm:"not null"`
	Phone       string `json:"phone" gorm:"not null"`
	Gender      string `json:"gender" gorm:"not null;co"`
	ParentName  string `json:"parent_name" gorm:"not null"`
	ParentPhone string `json:"parent_phone" gorm:"not null"`
	Apartment   string `json:"apartment" gorm:"not null"`
	ApartmentID int    `json:"apartment_id" gorm:"not null"`
	TeacherName string `json:"teacher_name" gorm:"not null"`
}

type ParamPassword struct {
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
	RePassword  string `json:"re_password"`
}

type ParamRecord struct {
	Name        string    `json:"name"`
	LeaveType   string    `json:"leave_type"`
	LeaveDays   float64   `json:"leave_days"`
	LeaveReason string    `json:"leave_reason"`
	LeaveSchool string    `json:"leave_school"`
	PlaceDetail string    `json:"place_detail"`
	LeaveWay    string    `json:"leave_way"`
	BackSchool  string    `json:"back_school"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Date        string    `json:"date"`
	Nigao       time.Time `json:"nigao"`
}

type ParamRecordQuery struct {
	RecordID int `json:"record_id"`
}
type VersionInfo struct {
	Latest      string   `json:"latest"`
	DownloadURL string   `json:"downloadURL"`
	UpdateLog   []string `json:"update_log"`
}
