package responses

type ErrorResponse struct {
	Message string `json:"message" example:"An error occurred."`
}
