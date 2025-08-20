package events

import (
	"context"
	"fmt"
	"otp-server/internal/infrastructure/config"
	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/redis"
)

type Publisher struct {
	redisClient *redis.Client
	config      *config.EventsConfig
	logger      logger.Logger
}

func NewPublisher(redisClient *redis.Client, cfg *config.EventsConfig, logger logger.Logger) *Publisher {
	return &Publisher{
		redisClient: redisClient,
		config:      cfg,
		logger:      logger,
	}
}

func (p *Publisher) Publish(ctx context.Context, event *Event) error {
	if !p.config.Enabled {
		return nil
	}

	if !p.isEventEnabled(event.Type) {
		return nil
	}

	data, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	err = p.redisClient.Publish(ctx, p.config.RedisChannel, string(data))
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (p *Publisher) PublishOTPGenerated(ctx context.Context, phoneNumber, otpCode string) error {
	event := NewEvent(p.config.EventTypes.OTPGenerated.Name, map[string]interface{}{
		"phone_number": phoneNumber,
		"otp_code":     otpCode,
	})
	return p.Publish(ctx, event)
}

func (p *Publisher) PublishOTPVerified(ctx context.Context, phoneNumber string, userID int) error {
	event := NewEvent(p.config.EventTypes.OTPVerified.Name, map[string]interface{}{
		"phone_number": phoneNumber,
		"user_id":      userID,
	})
	return p.Publish(ctx, event)
}

func (p *Publisher) PublishUserCreated(ctx context.Context, userID int, phoneNumber string) error {
	event := NewEvent(p.config.EventTypes.UserCreated.Name, map[string]interface{}{
		"user_id":      userID,
		"phone_number": phoneNumber,
	})
	return p.Publish(ctx, event)
}

func (p *Publisher) PublishUserLoggedIn(ctx context.Context, userID int, phoneNumber string) error {
	event := NewEvent(p.config.EventTypes.UserLoggedIn.Name, map[string]interface{}{
		"user_id":      userID,
		"phone_number": phoneNumber,
	})
	return p.Publish(ctx, event)
}

func (p *Publisher) PublishRateLimited(ctx context.Context, endpoint, identifier string) error {
	event := NewEvent(p.config.EventTypes.RateLimited.Name, map[string]interface{}{
		"endpoint":   endpoint,
		"identifier": identifier,
	})
	return p.Publish(ctx, event)
}

func (p *Publisher) isEventEnabled(eventType string) bool {
	switch eventType {
	case p.config.EventTypes.OTPGenerated.Name:
		return p.config.EventTypes.OTPGenerated.Enabled
	case p.config.EventTypes.OTPVerified.Name:
		return p.config.EventTypes.OTPVerified.Enabled
	case p.config.EventTypes.UserCreated.Name:
		return p.config.EventTypes.UserCreated.Enabled
	case p.config.EventTypes.UserLoggedIn.Name:
		return p.config.EventTypes.UserLoggedIn.Enabled
	case p.config.EventTypes.RateLimited.Name:
		return p.config.EventTypes.RateLimited.Enabled
	default:
		return true
	}
}
