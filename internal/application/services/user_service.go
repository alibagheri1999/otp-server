package services

import (
	"context"
	"fmt"
	"otp-server/internal/domain/entities"
	"otp-server/internal/domain/repositories"
	logger "otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/metrics"
	"otp-server/internal/infrastructure/redis"
	"strings"
)

type UserService struct {
	userRepo    repositories.UserRepository
	logger      logger.Logger
	redisClient *redis.Client
	cache       repositories.UserCacheRepository
	metrics     *metrics.MetricsService
}

func NewUserService(userRepo repositories.UserRepository, logger logger.Logger, redisClient *redis.Client, cacheRepo repositories.UserCacheRepository, metricsService *metrics.MetricsService) *UserService {
	return &UserService{
		userRepo:    userRepo,
		logger:      logger,
		redisClient: redisClient,
		cache:       cacheRepo,
		metrics:     metricsService,
	}
}

func (s *UserService) GetUserByID(ctx context.Context, userID int) (*entities.User, error) {
	user, err := s.cache.GetUserByID(ctx, userID)
	if err == nil {
		return user, nil
	}

	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.cache.SetUserByID(ctx, user); err != nil {
		s.logger.Error(ctx, "failed to cache user by ID", logger.F("userID", userID), logger.F("error", err))
	}

	return user, nil
}

func (s *UserService) IsAdmin(ctx context.Context, userID int) (bool, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return user.IsAdmin(), nil
}

func (s *UserService) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*entities.User, error) {
	return s.userRepo.GetByPhoneNumber(ctx, phoneNumber)
}

func (s *UserService) UpdateUserProfile(ctx context.Context, userID int, name string) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.UpdateProfile(name, user.PhoneNumber)

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	if err := s.cache.InvalidateUser(ctx, userID); err != nil {
		s.logger.Error(ctx, "failed to invalidate user cache", logger.F("userID", userID), logger.F("error", err))
	}

	return user, nil
}

func (s *UserService) UpdateLastSeen(ctx context.Context, userID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.UpdateLastSeen()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// GetUsers is a unified method that handles both search and pagination
func (s *UserService) GetUsers(ctx context.Context, query string, offset, limit int) ([]*entities.User, int, error) {
	users, total, err := s.cache.GetUsers(ctx, query, offset, limit)
	if err == nil {
		return users, total, nil
	}

	users, total, err = s.userRepo.GetUsersWithQuery(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	if err := s.cache.SetUsers(ctx, query, offset, limit, users, total); err != nil {
		s.logger.Error(ctx, "failed to cache unified users", logger.F("query", query), logger.F("offset", offset), logger.F("limit", limit), logger.F("error", err))
	}

	return users, total, nil
}

func (s *UserService) ActivateUser(ctx context.Context, userID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.Activate()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}

func (s *UserService) DeactivateUser(ctx context.Context, userID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	user.Deactivate()

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	return strings.Contains(sLower, substrLower)
}
