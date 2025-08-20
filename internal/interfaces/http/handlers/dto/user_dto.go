package dto

import "time"

// UpdateProfileRequest represents the request to update user profile
// @Description Request to update user profile information
type UpdateProfileRequest struct {
	// @Description User's full name
	// @Example John Doe
	// @Required
	Name string `json:"name" binding:"required" example:"John Doe"`
}

// UserResponse represents the user response
// @Description User information
type UserResponse struct {
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
	Role string `json:"role" example:"user"`
	// @Description Whether the user account is active
	// @Example true
	IsActive bool `json:"is_active" example:"true"`
	// @Description Account creation timestamp
	// @Example 2024-01-01T00:00:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	// @Description Last profile update timestamp
	// @Example 2024-01-01T00:00:00Z
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// UsersListResponse represents the response for getting users list
// @Description Response containing a list of users with pagination info
type UsersListResponse struct {
	// @Description List of users
	Users []UserResponse `json:"users"`
	// @Description Total number of users
	// @Example 100
	Total int `json:"total" example:"100"`
	// @Description Current offset
	// @Example 0
	Offset int `json:"offset" example:"0"`
	// @Description Current limit
	// @Example 10
	Limit int `json:"limit" example:"10"`
}

// UsersSearchResponse represents the response for searching users
// @Description Response containing search results for users
type UsersSearchResponse struct {
	// @Description List of matching users
	Users []UserResponse `json:"users"`
	// @Description Search query used
	// @Example john
	Query string `json:"query" example:"john"`
	// @Description Number of results found
	// @Example 5
	Count int `json:"count" example:"5"`
}

// UnifiedUsersRequest represents a request for users with optional search and pagination
type UnifiedUsersRequest struct {
	Query  string `json:"query" form:"query" binding:"omitempty"`     // Search query (optional)
	Offset int    `json:"offset" form:"offset" binding:"min=0"`       // Pagination offset
	Limit  int    `json:"limit" form:"limit" binding:"min=1,max=100"` // Pagination limit
}

// UnifiedUsersResponse represents the response for the unified users endpoint
type UnifiedUsersResponse struct {
	Users []*UserResponse `json:"users"`
	Total int             `json:"total"`
	Query string          `json:"query,omitempty"`
	Page  struct {
		Offset int `json:"offset"`
		Limit  int `json:"limit"`
	} `json:"page"`
}
