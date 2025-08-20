package dto

// SendOTPRequest represents the request to send OTP
// @Description Request to send OTP to a phone number
type SendOTPRequest struct {
	// @Description Phone number in international format (e.g., +1234567890)
	// @Example +1234567890
	// @Required
	PhoneNumber string `json:"phone_number" binding:"required" example:"+1234567890"`
}

// VerifyOTPRequest represents the request to verify OTP
// @Description Request to verify OTP and authenticate user
type VerifyOTPRequest struct {
	// @Description Phone number in international format
	// @Example +1234567890
	// @Required
	PhoneNumber string `json:"phone_number" binding:"required" example:"+1234567890"`
	// @Description One-time password (6 digits)
	// @Example 123456
	// @Required
	OTP string `json:"otp" binding:"required" example:"123456"`
	// @Description User's name (required for new user registration)
	// @Example John Doe
	// @Required
	Name string `json:"name" binding:"required" example:"John Doe"`
}

// AuthResponse represents the authentication response
// @Description Successful authentication response with token and user info
type AuthResponse struct {
	// @Description JWT token for API authentication
	// @Example eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// @Description User information
	User AuthUserResponse `json:"user"`
}

// AuthUserResponse represents the user response for authentication
// @Description User profile information for authentication
type AuthUserResponse struct {
	// @Description Unique user identifier
	// @Example 123
	ID int `json:"id" example:"123"`
	// @Description User's phone number
	// @Example +1234567890
	PhoneNumber string `json:"phone_number" example:"+1234567890"`
	// @Description User's full name
	// @Example John Doe
	Name string `json:"name" example:"John Doe"`
	// @Description User's role in the system
	// @Example user
	Role string `json:"role" example:"user" enums:"user,admin"`
}

// SendOTPResponse represents the response when OTP is sent successfully
// @Description Response when OTP is sent successfully
type SendOTPResponse struct {
	// @Description Success message
	// @Example OTP sent successfully
	Message string `json:"message" example:"OTP sent successfully"`
	// @Description Phone number where OTP was sent
	// @Example +1234567890
	PhoneNumber string `json:"phone_number" example:"+1234567890"`
	// @Description Timestamp when OTP was sent
	// @Example 2024-01-15T10:30:00Z
	Timestamp string `json:"timestamp" example:"2024-01-15T10:30:00Z"`
}

// RefreshTokenResponse represents the response when token is refreshed successfully
// @Description Response when access token is refreshed successfully
type RefreshTokenResponse struct {
	// @Description New JWT access token
	// @Example eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	// @Description Success message
	// @Example Token refreshed successfully
	Message string `json:"message" example:"Token refreshed successfully"`
	// @Description Token expiration time in seconds
	// @Example 3600
	ExpiresIn int `json:"expires_in" example:"3600"`
}
