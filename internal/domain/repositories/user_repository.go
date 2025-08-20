package repositories

import (
	"context"
	"otp-server/internal/domain/entities"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entities.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id int) (*entities.User, error)

	// GetByPhoneNumber retrieves a user by phone number
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*entities.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *entities.User) error

	// Delete deletes a user by ID
	Delete(ctx context.Context, id int) error

	// GetUsers retrieves a paginated list of users
	GetUsers(ctx context.Context, offset, limit int) ([]*entities.User, error)

	// GetTotalCount retrieves the total number of users
	GetTotalCount(ctx context.Context) (int, error)

	// SearchUsers searches users by phone number or name
	SearchUsers(ctx context.Context, query string) ([]*entities.User, error)

	// GetUsersWithQuery retrieves users with optional search and pagination in one query
	GetUsersWithQuery(ctx context.Context, query string, offset, limit int) ([]*entities.User, int, error)
}
