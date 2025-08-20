package middleware

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"otp-server/internal/infrastructure/config"
	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/metrics"
	"otp-server/internal/infrastructure/redis"
	"otp-server/internal/interfaces/http/handlers/dto"

	"github.com/gofiber/fiber/v2"
)

type RateLimiter struct {
	config      *config.Config
	logger      logger.Logger
	redisClient *redis.Client
	metrics     *metrics.MetricsService
}

type RateLimitMiddleware struct {
	rateLimiter *RateLimiter
}

func NewRateLimitMiddleware(cfg *config.Config, logger logger.Logger, redisClient *redis.Client, metricsService *metrics.MetricsService) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		rateLimiter: &RateLimiter{
			config:      cfg,
			logger:      logger,
			redisClient: redisClient,
			metrics:     metricsService,
		},
	}
}

func (rlm *RateLimitMiddleware) Global() fiber.Handler {
	return rlm.rateLimiter.GlobalRateLimit()
}

func (rlm *RateLimitMiddleware) Auth() fiber.Handler {
	return rlm.rateLimiter.AuthRateLimit()
}

func (rlm *RateLimitMiddleware) OTP() fiber.Handler {
	return rlm.rateLimiter.OTPRateLimit()
}

func (rlm *RateLimitMiddleware) User() fiber.Handler {
	return rlm.rateLimiter.UserRateLimit()
}

func (rlm *RateLimitMiddleware) AddRateLimitHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Next()

		clientIP := c.IP()
		endpointType := rlm.rateLimiter.getEndpointType(c.Path())

		// Add rate limit headers to all responses
		headers := rlm.rateLimiter.GetRateLimitHeaders(c.UserContext(), clientIP, endpointType)
		for key, value := range headers {
			c.Set(key, value)
		}

		// If this is a rate limit exceeded response (429), add additional headers
		if c.Response().StatusCode() == 429 {
			c.Set("Retry-After", headers["X-RateLimit-Reset"])
			c.Set("X-RateLimit-Exceeded", "true")
		}

		return nil
	}
}

func (rl *RateLimiter) GlobalRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIP := c.IP()
		key := fmt.Sprintf("rate_limit:global:%s", clientIP)

		if rl.config == nil || rl.config.RateLimiting.Global.Requests == 0 {
			rl.logger.Error(c.UserContext(), "Rate limiting config not properly initialized")
			return c.Next()
		}

		limit := rl.config.RateLimiting.Global.Requests
		duration := rl.config.RateLimiting.Global.Duration

		if err := rl.checkRateLimit(c.UserContext(), key, limit, duration, clientIP, "global"); err != nil {
			c.Set("Retry-After", strconv.FormatInt(int64(duration.Seconds()), 10))
			return c.Status(429).JSON(dto.ErrorResponse{
				Error:   "rate_limit_exceeded",
				Message: err.Error(),
			})
		}
		return c.Next()
	}
}

func (rl *RateLimiter) AuthRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIP := c.IP()
		key := fmt.Sprintf("rate_limit:auth:%s", clientIP)

		if rl.config == nil || rl.config.RateLimiting.Auth.Requests == 0 {
			rl.logger.Error(c.UserContext(), "Rate limiting config not properly initialized")
			return c.Next()
		}

		limit := rl.config.RateLimiting.Auth.Requests
		duration := rl.config.RateLimiting.Auth.Duration

		if err := rl.checkRateLimit(c.UserContext(), key, limit, duration, clientIP, "auth"); err != nil {
			c.Set("Retry-After", strconv.FormatInt(int64(duration.Seconds()), 10))
			return c.Status(429).JSON(dto.ErrorResponse{
				Error:   "rate_limit_exceeded",
				Message: err.Error(),
			})
		}
		return c.Next()
	}
}

func (rl *RateLimiter) OTPRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			PhoneNumber string `json:"phone_number"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		if rl.config == nil || rl.config.RateLimiting.OTP.Requests == 0 {
			rl.logger.Error(c.UserContext(), "Rate limiting config not properly initialized")
			return c.Next()
		}

		key := fmt.Sprintf("rate_limit:otp:%s", req.PhoneNumber)
		limit := rl.config.RateLimiting.OTP.Requests
		duration := rl.config.RateLimiting.OTP.Duration

		if err := rl.checkRateLimit(c.UserContext(), key, limit, duration, req.PhoneNumber, "otp"); err != nil {
			c.Set("Retry-After", strconv.FormatInt(int64(duration.Seconds()), 10))
			return c.Status(429).JSON(dto.ErrorResponse{
				Error:   "rate_limit_exceeded",
				Message: err.Error(),
			})
		}
		return c.Next()
	}
}

func (rl *RateLimiter) UserRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIP := c.IP()
		key := fmt.Sprintf("rate_limit:user:%s", clientIP)

		if rl.config == nil || rl.config.RateLimiting.User.Requests == 0 {
			rl.logger.Error(c.UserContext(), "Rate limiting config not properly initialized")
			return c.Next()
		}

		limit := rl.config.RateLimiting.User.Requests
		duration := rl.config.RateLimiting.User.Duration

		if err := rl.checkRateLimit(c.UserContext(), key, limit, duration, clientIP, "user"); err != nil {
			c.Set("Retry-After", strconv.FormatInt(int64(duration.Seconds()), 10))
			return c.Status(429).JSON(dto.ErrorResponse{
				Error:   "rate_limit_exceeded",
				Message: err.Error(),
			})
		}
		return c.Next()
	}
}

func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string, limit int, duration time.Duration, identifier, endpointType string) error {
	current, err := rl.redisClient.Get(ctx, key)
	if err != nil && current != "" {
		rl.logger.Error(ctx, "Failed to get rate limit", logger.F("error", err), logger.F("key", key))
	}

	var count int
	if current != "" {
		count, _ = strconv.Atoi(current)
	}

	if count >= limit {
		rl.logger.Warn(ctx, "Rate limit exceeded",
			logger.F("endpoint_type", endpointType),
			logger.F("identifier", identifier),
			logger.F("limit", limit),
			logger.F("duration", duration))

		if rl.metrics != nil {
			rl.metrics.RecordRateLimitExceeded(endpointType, identifier)
		}

		return fmt.Errorf("too many requests. Limit: %d requests per %v. Please try again later.", limit, duration)
	}

	count++
	err = rl.redisClient.Set(ctx, key, strconv.Itoa(count), duration)
	if err != nil {
		rl.logger.Error(ctx, "Failed to set rate limit", logger.F("error", err), logger.F("key", key))
	}

	return nil
}

func (rl *RateLimiter) GetRateLimitHeaders(ctx context.Context, identifier, endpointType string) map[string]string {
	headers := make(map[string]string)

	if rl.config == nil {
		return headers
	}

	switch endpointType {
	case "global":
		if rl.config.RateLimiting.Global.Requests > 0 {
			headers["X-RateLimit-Limit"] = strconv.Itoa(rl.config.RateLimiting.Global.Requests)
			headers["X-RateLimit-Remaining"] = rl.getRemainingRequests(ctx, fmt.Sprintf("rate_limit:global:%s", identifier), rl.config.RateLimiting.Global.Requests)
			headers["X-RateLimit-Reset"] = rl.getResetTime(ctx, fmt.Sprintf("rate_limit:global:%s", identifier))
		}
	case "auth":
		if rl.config.RateLimiting.Auth.Requests > 0 {
			headers["X-RateLimit-Limit"] = strconv.Itoa(rl.config.RateLimiting.Auth.Requests)
			headers["X-RateLimit-Remaining"] = rl.getRemainingRequests(ctx, fmt.Sprintf("rate_limit:auth:%s", identifier), rl.config.RateLimiting.Auth.Requests)
			headers["X-RateLimit-Reset"] = rl.getResetTime(ctx, fmt.Sprintf("rate_limit:auth:%s", identifier))
		}
	case "otp":
		if rl.config.RateLimiting.OTP.Requests > 0 {
			headers["X-RateLimit-Limit"] = strconv.Itoa(rl.config.RateLimiting.OTP.Requests)
			headers["X-RateLimit-Remaining"] = rl.getRemainingRequests(ctx, fmt.Sprintf("rate_limit:otp:%s", identifier), rl.config.RateLimiting.OTP.Requests)
			headers["X-RateLimit-Reset"] = rl.getResetTime(ctx, fmt.Sprintf("rate_limit:otp:%s", identifier))
		}
	case "user":
		if rl.config.RateLimiting.User.Requests > 0 {
			headers["X-RateLimit-Limit"] = strconv.Itoa(rl.config.RateLimiting.User.Requests)
			headers["X-RateLimit-Remaining"] = rl.getRemainingRequests(ctx, fmt.Sprintf("rate_limit:user:%s", identifier), rl.config.RateLimiting.User.Requests)
			headers["X-RateLimit-Reset"] = rl.getResetTime(ctx, fmt.Sprintf("rate_limit:user:%s", identifier))
		}
	}

	return headers
}

func (rl *RateLimiter) getRemainingRequests(ctx context.Context, key string, limit int) string {
	current, err := rl.redisClient.Get(ctx, key)
	if err != nil || current == "" {
		return strconv.Itoa(limit)
	}

	count, _ := strconv.Atoi(current)
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}

	return strconv.Itoa(remaining)
}

func (rl *RateLimiter) getResetTime(ctx context.Context, key string) string {
	ttl, err := rl.redisClient.TTL(ctx, key)
	if err != nil {
		return "0"
	}

	return strconv.FormatInt(int64(ttl.Seconds()), 10)
}

func (rl *RateLimiter) getEndpointType(path string) string {
	if strings.Contains(path, "/auth/send-otp") {
		return "otp"
	}
	if strings.Contains(path, "/auth/") {
		return "auth"
	}
	if strings.Contains(path, "/users/") {
		return "user"
	}
	return "global"
}
