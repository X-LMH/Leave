package dto

// AdminLoginRequest is the JSON body accepted by the management platform login endpoint.
type AdminLoginRequest struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLoginResponse contains the token used by the management platform.
type AdminLoginResponse struct {
	Token string `json:"token"`
}

// AuthIdentity is the current, database-verified management identity.
type AuthIdentity struct {
	StudentID string `json:"student_id"`
	Role      string `json:"role"`
}
