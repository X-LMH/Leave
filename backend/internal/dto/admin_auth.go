package dto

// AdminLoginRequest is the JSON body accepted by the management platform login endpoint.
type AdminLoginRequest struct {
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLoginResponse contains the token used by the management platform.
type AdminLoginResponse struct {
	Token string `json:"token"`
}
