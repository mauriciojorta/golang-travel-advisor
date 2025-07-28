package requests

type SignUpRequest struct {
	Email    string `json:"email" binding:"required,max=128" example:"test@example.com"`
	Password string `json:"password" binding:"required,max=256" example:"Password123-"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,max=100" example:"test@example.com"`
	Password string `json:"password" binding:"required,max=256" example:"Password123-"`
}
