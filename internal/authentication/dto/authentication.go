package dto

type SignUpRequest struct {
	Data SignUpData `json:"data" binding:"required"`
}

type SignUpData struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Username string `json:"username" binding:"required,min=8"`
}

type SignUpResponse struct {
	AccessToken string `json:"access_token"`
}
