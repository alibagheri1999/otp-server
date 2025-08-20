package database

import (
	"otp-server/internal/domain/repositories"
)

// Repositories holds all repository interfaces
type Repositories struct {
	UserRepository      repositories.UserRepository
	UserCacheRepository repositories.UserCacheRepository
}

// NewRepositories creates a new repositories instance
func NewRepositories(postgresPool *PostgresPool, redisClient interface{}) *Repositories {
	return &Repositories{
		UserRepository:      NewUserRepository(postgresPool),
		UserCacheRepository: nil,
	}
}

// SetUserCacheRepository sets the user cache repository
func (r *Repositories) SetUserCacheRepository(cacheRepo repositories.UserCacheRepository) {
	r.UserCacheRepository = cacheRepo
}
