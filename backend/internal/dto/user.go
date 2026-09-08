// Package dto contains request and response shapes exposed by the HTTP API.
package dto

// RegisterRequest is the JSON body accepted by the registration endpoint.
type RegisterRequest struct {
	StudentID  string `json:"student_id" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RePassword string `json:"re_password" binding:"required"`
}

// LoginRequest is the JSON body accepted by the login endpoint.
type LoginRequest struct {
	StudentID  string `json:"student_id" binding:"required"`
	Password   string `json:"password" binding:"required"`
	DeviceName string `json:"device_name" binding:"max=255"`
	AppVersion string `json:"app_version" binding:"max=64"`
}

// User contains the basic user data needed by authenticated pages.
type User struct {
	StudentID string `json:"student_id"`
	Name      string `json:"name"`
	ClassName string `json:"class_name"`
}

// LoginResponse is returned in the data field after a successful login.
type LoginResponse struct {
	Token            string `json:"token"`
	User             User   `json:"user"`
	ProfileCompleted bool   `json:"profile_completed"`
}

// PasswordChangeRequest is the JSON body accepted by the password endpoint.
type PasswordChangeRequest struct {
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
	RePassword  string `json:"re_password"`
}
