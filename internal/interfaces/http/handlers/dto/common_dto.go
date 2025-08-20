package dto

// ErrorResponse represents an error response
// @Description Standard error response format
type ErrorResponse struct {
	// @Description Error type or category
	// @Example validation_error
	Error string `json:"error" example:"validation_error"`
	// @Description Detailed error message
	// @Example Phone number format is invalid
	Message string `json:"message" example:"Phone number format is invalid"`
}
