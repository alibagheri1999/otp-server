package events

import (
	"context"
	"otp-server/internal/infrastructure/config"
	"otp-server/internal/infrastructure/logger"
	"otp-server/internal/infrastructure/redis"
)

type EventService struct {
	publisher  *Publisher
	subscriber *Subscriber
	logger     logger.Logger
}

func NewEventService(redisClient *redis.Client, cfg *config.EventsConfig, logger logger.Logger) *EventService {
	return &EventService{
		publisher:  NewPublisher(redisClient, cfg, logger),
		subscriber: NewSubscriber(redisClient, cfg, logger),
		logger:     logger,
	}
}

func (es *EventService) Publish(ctx context.Context, event *Event) error {
	return es.publisher.Publish(ctx, event)
}

func (es *EventService) PublishOTPGenerated(ctx context.Context, phoneNumber, otpCode string) error {
	return es.publisher.PublishOTPGenerated(ctx, phoneNumber, otpCode)
}

func (es *EventService) PublishOTPVerified(ctx context.Context, phoneNumber string, userID int) error {
	return es.publisher.PublishOTPVerified(ctx, phoneNumber, userID)
}

func (es *EventService) PublishUserCreated(ctx context.Context, userID int, phoneNumber string) error {
	return es.publisher.PublishUserCreated(ctx, userID, phoneNumber)
}

func (es *EventService) PublishUserLoggedIn(ctx context.Context, userID int, phoneNumber string) error {
	return es.publisher.PublishUserLoggedIn(ctx, userID, phoneNumber)
}

func (es *EventService) PublishRateLimited(ctx context.Context, endpoint, identifier string) error {
	return es.publisher.PublishRateLimited(ctx, endpoint, identifier)
}

func (es *EventService) Subscribe(ctx context.Context, eventType string, handler EventHandler) error {
	return es.subscriber.Subscribe(ctx, eventType, handler)
}

func (es *EventService) Start(ctx context.Context) error {
	return es.subscriber.Start(ctx)
}

func (es *EventService) GetPublisher() *Publisher {
	return es.publisher
}

func (es *EventService) GetSubscriber() *Subscriber {
	return es.subscriber
}
