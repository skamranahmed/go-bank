package authentication

type signUpRequest struct {
	Data struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
		Username string `json:"username" binding:"required,min=8"`
	} `json:"data" binding:"required"`
}

type signUpResponse struct {
	AccessToken string `json:"access_token"`
}
