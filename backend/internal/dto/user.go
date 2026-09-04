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
	StudentID string `json:"student_id" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// LoginResponse is returned in the data field after a successful login.
type LoginResponse struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}
