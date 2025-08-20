package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"otp-server/internal/application"
	"otp-server/internal/infrastructure/config"
	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/metrics"
	"otp-server/internal/infrastructure/redis"

	"github.com/gofiber/fiber/v2"
)

// Middleware holds all middleware functions
type Middleware struct {
	authService application.AuthServiceInterface
	config      *config.Config
	logger      logger.Logger
	redisClient *redis.Client
	metrics     *metrics.MetricsService
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(config *config.Config, logger logger.Logger, redisClient *redis.Client) *Middleware {
	return &Middleware{
		config:      config,
		logger:      logger,
		redisClient: redisClient,
	}
}

// SetAuthService sets the auth service for middleware
func (m *Middleware) SetAuthService(authService application.AuthServiceInterface) {
	m.authService = authService
}

// SetMetricsService sets the metrics service for middleware
func (m *Middleware) SetMetricsService(metricsService *metrics.MetricsService) {
	m.metrics = metricsService
}

// GetLogger returns the logger instance
func (m *Middleware) GetLogger() logger.Logger {
	return m.logger
}

// GetRedisClient returns the Redis client instance
func (m *Middleware) GetRedisClient() *redis.Client {
	return m.redisClient
}

// GetMetricsService returns the metrics service instance
func (m *Middleware) GetMetricsService() *metrics.MetricsService {
	return m.metrics
}

// CORS middleware for handling Cross-Origin Resource Sharing
func (m *Middleware) CORS() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if c.Method() == http.MethodOptions {
			return c.SendStatus(http.StatusNoContent)
		}

		return c.Next()
	}
}

// SecurityHeaders middleware for setting security headers
func (m *Middleware) SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		return c.Next()
	}
}

// Auth middleware for JWT token authentication
func (m *Middleware) Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(map[string]interface{}{
				"error":   "Authorization header required",
				"message": "Please provide a valid authorization header",
			})
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			return c.Status(http.StatusUnauthorized).JSON(map[string]interface{}{
				"error":   "Invalid authorization header",
				"message": "Authorization header must start with 'Bearer '",
			})
		}

		tokenString := authHeader[7:]

		user, err := m.authService.GetUserFromToken(tokenString)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(map[string]interface{}{
				"error":   "Invalid token",
				"message": err.Error(),
			})
		}

		c.Locals("user", user)
		c.Locals("user_id", user.ID)

		if uc := c.UserContext(); uc != nil {
			c.SetUserContext(context.WithValue(uc, "user", user))
		}

		return c.Next()
	}
}

// RateLimit middleware for rate limiting requests using Redis
func (m *Middleware) RateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIP := c.IP()
		key := "rate_limit:" + clientIP

		current, err := m.redisClient.Get(c.UserContext(), key)
		if err != nil && err.Error() != "redis: nil" {
			m.logger.Error(c.UserContext(), "Rate limit check failed", logger.F("error", err))
			return c.Next()
		}

		var count int64
		if current != "" {
			count, _ = strconv.ParseInt(current, 10, 64)
		}

		if count >= 100 {
			m.logger.Warn(c.UserContext(), "Rate limit exceeded",
				logger.F("client_ip", clientIP),
				logger.F("count", count))
			return c.Status(http.StatusTooManyRequests).JSON(map[string]interface{}{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests, please try again later",
				"retry_after": 60,
			})
		}

		pipe := m.redisClient.GetClient().Pipeline()
		pipe.Incr(c.UserContext(), key)
		pipe.Expire(c.UserContext(), key, time.Minute)
		_, err = pipe.Exec(c.UserContext())

		if err != nil {
			m.logger.Error(c.UserContext(), "Rate limit update failed", logger.F("error", err))
		}

		return c.Next()
	}
}

// Logging middleware for request logging
func (m *Middleware) Logging() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		latency := time.Since(start)
		clientIP := c.IP()
		method := c.Method()
		path := c.OriginalURL()
		statusCode := c.Response().StatusCode()

		if m.metrics != nil {
			m.metrics.RecordRequest(method, path, statusCode, latency)
		}

		m.logger.Info(c.UserContext(), "HTTP Request",
			logger.F("method", method),
			logger.F("path", path),
			logger.F("status", statusCode),
			logger.F("latency", latency),
			logger.F("client_ip", clientIP),
		)

		return err
	}
}

// ErrorHandler middleware for handling panics
func (m *Middleware) ErrorHandler() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if rec := recover(); rec != nil {
				m.logger.Error(c.UserContext(), "Panic recovered", logger.F("error", rec))
				_ = c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
					"error":   "Internal server error",
					"message": "Something went wrong",
				})
			}
		}()
		return c.Next()
	}
}
