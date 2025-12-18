package api

type ErrorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}

type MessageResponse struct {
	Message string `json:"message" example:"ok"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}
