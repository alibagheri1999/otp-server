package repositories

import (
	"context"

	"otp-server/internal/domain/entities"
)

// UserCacheRepository defines the interface for user caching operations
type UserCacheRepository interface {
	// GetUserByID retrieves a user from cache by ID
	GetUserByID(ctx context.Context, userID int) (*entities.User, error)

	// SetUserByID stores a user in cache by ID
	SetUserByID(ctx context.Context, user *entities.User) error

	// GetUserByPhoneNumber retrieves a user from cache by phone number
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*entities.User, error)

	// SetUserByPhoneNumber stores a user in cache by phone number
	SetUserByPhoneNumber(ctx context.Context, user *entities.User) error

	// GetUsers retrieves users from cache with optional search and pagination
	GetUsers(ctx context.Context, query string, offset, limit int) ([]*entities.User, int, error)

	// SetUsers stores users in cache with optional search and pagination
	SetUsers(ctx context.Context, query string, offset, limit int, users []*entities.User, total int) error

	// InvalidateUser removes all cached data for a specific user
	InvalidateUser(ctx context.Context, userID int) error

	// InvalidateAll removes all cached user data
	InvalidateAll(ctx context.Context) error
}
