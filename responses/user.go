package responses

type SignUpResponse struct {
	Message string `json:"message" example:"User created."`
	User    string `json:"user" example:"test@example.com"`
}

type LoginResponse struct {
	Message string `json:"message" example:"Login successful!"`
	Token   string `json:"token" example:"token123"`
}
