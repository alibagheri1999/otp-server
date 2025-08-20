package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"otp-server/internal/domain/entities"
	"otp-server/internal/domain/repositories"
	logger "otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/metrics"
	"otp-server/internal/infrastructure/redis"
)

// UserCacheService implements the UserCacheRepository interface
type UserCacheService struct {
	redisClient *redis.Client
	logger      logger.Logger
	ttl         time.Duration
	metrics     *metrics.MetricsService
}

// Ensure UserCacheService implements UserCacheRepository interface
var _ repositories.UserCacheRepository = (*UserCacheService)(nil)

func NewUserCacheService(redisClient *redis.Client, logger logger.Logger, metricsService *metrics.MetricsService) *UserCacheService {
	return &UserCacheService{
		redisClient: redisClient,
		logger:      logger,
		ttl:         15 * time.Minute,
		metrics:     metricsService,
	}
}

func (c *UserCacheService) GetUserByID(ctx context.Context, userID int) (*entities.User, error) {
	key := fmt.Sprintf("user:id:%d", userID)

	data, err := c.redisClient.Get(ctx, key)
	if err != nil || data == "" {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss("user", key)
		}
		return nil, fmt.Errorf("user not found in cache")
	}

	if c.metrics != nil {
		c.metrics.RecordCacheHit("user", key)
	}

	var user entities.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *UserCacheService) SetUserByID(ctx context.Context, user *entities.User) error {
	key := fmt.Sprintf("user:id:%d", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	return c.redisClient.Set(ctx, key, string(data), c.ttl)
}

func (c *UserCacheService) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*entities.User, error) {
	key := fmt.Sprintf("user:phone:%s", phoneNumber)

	data, err := c.redisClient.Get(ctx, key)
	if err != nil || data == "" {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss("user", key)
		}
		return nil, fmt.Errorf("user not found in cache")
	}

	if c.metrics != nil {
		c.metrics.RecordCacheHit("user", key)
	}

	var user entities.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *UserCacheService) SetUserByPhoneNumber(ctx context.Context, user *entities.User) error {
	key := fmt.Sprintf("user:phone:%s", user.PhoneNumber)

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	return c.redisClient.Set(ctx, key, string(data), c.ttl)
}

func (c *UserCacheService) GetUsers(ctx context.Context, query string, offset, limit int) ([]*entities.User, int, error) {
	var key string
	if query != "" {
		key = fmt.Sprintf("users:search:%s:%d:%d", query, offset, limit)
	} else {
		key = fmt.Sprintf("users:list:%d:%d", offset, limit)
	}

	data, err := c.redisClient.Get(ctx, key)
	if err != nil || data == "" {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss("user", key)
		}
		return nil, 0, fmt.Errorf("users not found in cache")
	}

	if c.metrics != nil {
		c.metrics.RecordCacheHit("user", key)
	}

	var result struct {
		Users []*entities.User `json:"users"`
		Total int              `json:"total"`
	}

	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, 0, err
	}

	return result.Users, result.Total, nil
}

func (c *UserCacheService) SetUsers(ctx context.Context, query string, offset, limit int, users []*entities.User, total int) error {
	var key string
	if query != "" {
		key = fmt.Sprintf("users:search:%s:%d:%d", query, offset, limit)
	} else {
		key = fmt.Sprintf("users:list:%d:%d", offset, limit)
	}

	result := struct {
		Users []*entities.User `json:"users"`
		Total int              `json:"total"`
	}{
		Users: users,
		Total: total,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}

	return c.redisClient.Set(ctx, key, string(data), c.ttl)
}

func (c *UserCacheService) InvalidateUser(ctx context.Context, userID int) error {
	patterns := []string{
		fmt.Sprintf("user:id:%d", userID),
		"user:phone:*",
		"users:list:*",
		"users:search:*",
	}

	for _, pattern := range patterns {
		if err := c.redisClient.DelPattern(ctx, pattern); err != nil {
			c.logger.Error(ctx, "failed to invalidate cache pattern", logger.F("pattern", pattern), logger.F("error", err))
		}
	}

	return nil
}

func (c *UserCacheService) InvalidateAll(ctx context.Context) error {
	patterns := []string{
		"user:*",
		"users:*",
	}

	for _, pattern := range patterns {
		if err := c.redisClient.DelPattern(ctx, pattern); err != nil {
		}
	}

	return nil
}
