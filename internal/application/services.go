package application

import (
	"context"
	"otp-server/internal/infrastructure/config"
	log "otp-server/internal/infrastructure/logger"

	"otp-server/internal/application/services"
	"otp-server/internal/domain/entities"
	"otp-server/internal/infrastructure/cache"
	"otp-server/internal/infrastructure/database"
	"otp-server/internal/infrastructure/events"
	"otp-server/internal/infrastructure/metrics"
	"otp-server/internal/infrastructure/redis"
)

// Service interfaces
type AuthServiceInterface interface {
	SendOTP(ctx context.Context, phoneNumber string) error
	VerifyOTPAndAuthenticate(ctx context.Context, phoneNumber, otpCode, name string) (*entities.User, string, error)
	GetUserFromToken(tokenString string) (*entities.User, error)
}

type UserServiceInterface interface {
	GetUserByID(ctx context.Context, userID int) (*entities.User, error)
	GetUsers(ctx context.Context, query string, offset, limit int) ([]*entities.User, int, error)
	UpdateUserProfile(ctx context.Context, userID int, name string) (*entities.User, error)
}

// Services holds all application services
type Services struct {
	AuthService      AuthServiceInterface
	UserService      UserServiceInterface
	EventService     *events.EventService
	UserCacheService *cache.UserCacheService
}

// NewServices creates a new services container
func NewServices(repos *database.Repositories, config *config.Config, redisClient *redis.Client, metricsService *metrics.MetricsService) *Services {
	logger := log.New(config.Log)

	eventService := events.NewEventService(redisClient, &config.Events, logger)

	otpService := redis.NewOTPService(redisClient, &config.OTP, logger, metricsService)

	otpService.SetEventHandler(func(ctx context.Context, phoneNumber, otpCode string) error {
		return eventService.PublishOTPGenerated(ctx, phoneNumber, otpCode)
	})

	userCacheService := cache.NewUserCacheService(redisClient, logger, metricsService)

	repos.SetUserCacheRepository(userCacheService)

	return &Services{
		AuthService:      services.NewAuthService(repos.UserRepository, otpService, logger, config.JWT.Secret, metricsService),
		UserService:      services.NewUserService(repos.UserRepository, logger, redisClient, userCacheService, metricsService),
		EventService:     eventService,
		UserCacheService: userCacheService,
	}
}

// GetEventService returns the event service
func (s *Services) GetEventService() *events.EventService {
	return s.EventService
}

// GetUserCacheService returns the user cache service
func (s *Services) GetUserCacheService() *cache.UserCacheService {
	return s.UserCacheService
}
